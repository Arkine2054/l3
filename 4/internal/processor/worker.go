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
			idStr := string(m.Key)
			log.Println("processing image id:", idStr)
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

	err = w.Repo.UpdatePathsAndStatus(id, nil, nil, "processing")
	if err != nil {
		log.Println("update paths and status processing error:", err)
	}

	imageRecord, err := w.Repo.GetByID(id)
	if err != nil {
		log.Println("image not found:", err)
		return
	}

	img, err := imaging.Open(*imageRecord.StoredPath)
	if err != nil {
		log.Println("open error:", err)
		err = w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Println("update paths and status failed error:", err)
		}
		return
	}

	resized := imaging.Fit(img, 1200, 1200, imaging.Lanczos)

	if w.WatermarkText != "" {
		resized = addTextWatermark(resized, w.WatermarkText)
	}

	procDir := filepath.Join(w.StorageDir, "processed")

	err = os.MkdirAll(procDir, 0o755)
	if err != nil {
		log.Println("mkdir error:", err)
	}
	procPath := filepath.Join(procDir, filepath.Base(*imageRecord.StoredPath))
	if err := imaging.Save(resized, procPath); err != nil {
		log.Println("save processed error:", err)
		err = w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Println("update paths and status failed error:", err)
		}
		return
	}

	thumb := imaging.Thumbnail(resized, 300, 300, imaging.CatmullRom)
	thumbDir := filepath.Join(w.StorageDir, "thumbs")
	err = os.MkdirAll(thumbDir, 0o755)
	if err != nil {
		log.Println("mkdir error:", err)
	}
	thumbPath := filepath.Join(thumbDir, filepath.Base(*imageRecord.StoredPath))
	if err := imaging.Save(thumb, thumbPath); err != nil {
		log.Println("save thumb error:", err)
		err = w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Println("update paths and status thumbnail failed error:", err)
		}
		return
	}

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
		log.Println("Watermark: font not found:", err)
		return rgba
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println("Watermark: font parse error:", err)
		return rgba
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(float64(w) / 15)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 200}))

	pt := freetype.Pt(20, h-40)

	shadowCtx := freetype.NewContext()
	shadowCtx.SetDPI(72)
	shadowCtx.SetFont(f)
	shadowCtx.SetFontSize(float64(w) / 15)
	shadowCtx.SetClip(rgba.Bounds())
	shadowCtx.SetDst(rgba)
	shadowCtx.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 150}))
	_, err = shadowCtx.DrawString(text, freetype.Pt(int(pt.X+2), int(pt.Y+2)))
	if err != nil {
		log.Println("shadow error:", err)
	}

	_, err = c.DrawString(text, pt)
	if err != nil {
		log.Println("Watermark: draw string error:", err)
	} else {
		log.Println("Watermark visible on image")
	}

	return rgba
}
