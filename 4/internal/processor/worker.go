package processor

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

type Worker struct {
	Repo          ImageRepository
	Consumer      MessageConsumer
	StorageDir    string
	WatermarkText string
	FontPath      string
}

func (w *Worker) Run(ctx context.Context) {

	for {
		msg, err := w.Consumer.Read(ctx)
		if err != nil {
			log.Println("Worker: consumer error:", err)
			continue
		}
		w.handleMessage(ctx, msg)
	}
}

func (w *Worker) handleMessage(ctx context.Context, msg string) {

	var id int64
	if _, err := fmt.Sscanf(msg, "%d", &id); err != nil {
		log.Println("Worker: invalid message, cannot parse ID:", msg, "error:", err)
		return
	}

	if w.Repo == nil {
		log.Println("Worker: Repo is nil, cannot process")
		return
	}

	imgRec, err := w.Repo.GetByID(id)
	if err != nil {
		log.Println("Worker: image not found in DB, id:", id, "error:", err)
		return
	}

	if imgRec.StoredPath == nil {
		err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Printf("Worker: failed to update stored path for image id:%d, err:%v", id, err)
		}
		return
	}

	if err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "processing"); err != nil {
		log.Println("Worker: failed to update status to processing:", err)
		return
	}

	img, err := imaging.Open(*imgRec.StoredPath)
	if err != nil {
		err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Printf("Worker: failed to update stored path for image id:%d, err:%v", id, err)
		}
		return
	}

	resized := imaging.Fit(img, 1200, 1200, imaging.Lanczos)

	if w.WatermarkText != "" {
		resized = AddTextWatermark(resized, w.WatermarkText, w.FontPath)
	}

	procDir := filepath.Join(w.StorageDir, "processed")
	thumbDir := filepath.Join(w.StorageDir, "thumbs")
	if err := os.MkdirAll(procDir, 0o755); err != nil {
		err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Printf("Worker: failed to update procDir stored path for image id:%d, err:%v", id, err)
		}
		return
	}
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Printf("Worker: failed to update thumbDir stored path for image id:%d, err:%v", id, err)
		}
		return
	}

	procPath := filepath.Join(procDir, filepath.Base(*imgRec.StoredPath))
	if err := imaging.Save(resized, procPath); err != nil {
		err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Printf("Worker: failed to update procPath stored path for image id:%d, err:%v", id, err)
		}
		return
	}

	thumbPath := filepath.Join(thumbDir, filepath.Base(*imgRec.StoredPath))
	thumb := imaging.Thumbnail(resized, 300, 300, imaging.CatmullRom)
	if err := imaging.Save(thumb, thumbPath); err != nil {
		err := w.Repo.UpdatePathsAndStatus(id, nil, nil, "failed")
		if err != nil {
			log.Printf("Worker: failed to update thumbPath stored path for image id:%d, err:%v", id, err)
		}
		return
	}

	if err := w.Repo.UpdatePathsAndStatus(id, &procPath, &thumbPath, "done"); err != nil {
		log.Println("Worker: failed to update DB with paths:", err)
		return
	}
}

func (w *Worker) saveImages(img image.Image, origPath string) (string, string, error) {
	procDir := filepath.Join(w.StorageDir, "processed")
	thumbDir := filepath.Join(w.StorageDir, "thumbs")

	if err := os.MkdirAll(procDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir processed failed: %w", err)
	}
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir thumbs failed: %w", err)
	}

	name := filepath.Base(origPath)
	procPath := filepath.Join(procDir, name)
	thumbPath := filepath.Join(thumbDir, name)

	if err := imaging.Save(img, procPath); err != nil {
		return "", "", fmt.Errorf("save processed failed: %w", err)
	}

	thumb := imaging.Thumbnail(img, 300, 300, imaging.CatmullRom)
	if err := imaging.Save(thumb, thumbPath); err != nil {
		return "", "", fmt.Errorf("save thumbnail failed: %w", err)
	}

	return procPath, thumbPath, nil
}
