package processor

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/golang/freetype"

	"github.com/disintegration/imaging"
	"github.com/segmentio/kafka-go"
	"gitlab.com/arkine/l3/4/internal/repository"
)

type Worker struct {
	Repo          *repository.ImagesRepo
	Reader        *kafka.Reader
	StorageDir    string
	WatermarkText string
}

func (w *Worker) Start(ctx context.Context) {

	go func() {
		for {
			m, err := w.Reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Println("kafka read error:", err)
				continue
			}
			// key is id
			idStr := string(m.Key)
			log.Println("processing image id:", idStr)
			// process in a goroutine
			go w.handleMessage(ctx, idStr, string(m.Value))
		}
	}()
}

func (w *Worker) handleMessage(ctx context.Context, _ string, value string) {
	idStr := value
	log.Println("Processing image ID:", idStr, "WatermarkText:", w.WatermarkText)

	var id int64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		log.Println("invalid image ID:", err)
		return
	}

	// помечаем запись как processing
	_ = w.Repo.UpdatePathsAndStatus(id, nil, nil, "processing")

	// получаем запись из базы
	imageRecord, err := w.Repo.GetByID(id)
	if err != nil {
		log.Println("image not found:", err)
		return
	}

	// открываем исходное изображение
	img, err := imaging.Open(*imageRecord.StoredPath)
	if err != nil {
		log.Println("open error:", err)
		_ = w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		return
	}

	// ресайз до max 1200x1200
	resized := imaging.Fit(img, 1200, 1200, imaging.Lanczos)

	// добавляем watermark
	if w.WatermarkText != "" {
		resized = addTextWatermark(resized, w.WatermarkText)
	}

	// сохраняем обработанное изображение
	procDir := filepath.Join(w.StorageDir, "processed")
	os.MkdirAll(procDir, 0o755)
	procPath := filepath.Join(procDir, filepath.Base(*imageRecord.StoredPath))
	if err := imaging.Save(resized, procPath); err != nil {
		log.Println("save processed error:", err)
		_ = w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		return
	}

	// создаём thumbnail
	thumb := imaging.Thumbnail(resized, 300, 300, imaging.CatmullRom)
	thumbDir := filepath.Join(w.StorageDir, "thumbs")
	os.MkdirAll(thumbDir, 0o755)
	thumbPath := filepath.Join(thumbDir, filepath.Base(*imageRecord.StoredPath))
	if err := imaging.Save(thumb, thumbPath); err != nil {
		log.Println("save thumb error:", err)
		_ = w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		return
	}

	// обновляем пути и статус в базе
	p := procPath
	t := thumbPath
	if err := w.Repo.UpdatePathsAndStatus(id, &p, &t, "done"); err != nil {
		log.Println("db update error:", err)
	}
}

func addTextWatermark(img image.Image, text string) *image.NRGBA {
	rgba := imaging.Clone(img)
	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()

	fontBytes, err := os.ReadFile("/app/internal/assets/font.ttf")
	if err != nil {
		log.Println("⚠️ watermark: font not found:", err)
		return rgba
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println("⚠️ watermark: font parse error:", err)
		return rgba
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(float64(w) / 15) // крупнее шрифт
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 200})) // белый с небольшой прозрачностью

	pt := freetype.Pt(20, h-40)

	// Нарисуем лёгкую тень для читаемости
	shadowCtx := freetype.NewContext()
	shadowCtx.SetDPI(72)
	shadowCtx.SetFont(f)
	shadowCtx.SetFontSize(float64(w) / 15)
	shadowCtx.SetClip(rgba.Bounds())
	shadowCtx.SetDst(rgba)
	shadowCtx.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 150})) // чёрная полупрозрачная тень
	_, err = shadowCtx.DrawString(text, freetype.Pt(int(pt.X+2), int(pt.Y+2)))
	if err != nil {
		log.Println("shadow error:", err)
	}

	// Сам текст поверх
	_, err = c.DrawString(text, pt)
	if err != nil {
		log.Println("⚠️ watermark: draw string error:", err)
	} else {
		log.Println("✅ Watermark visible on image")
	}

	return rgba
}
