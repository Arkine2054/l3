package handlers

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"

	kafka2 "gitlab.com/arkine/l3/4/internal/kafka"
	"gitlab.com/arkine/l3/4/internal/repository"
)

type UploadHandler struct {
	Repo        *repository.ImagesRepo
	KafkaWriter *kafka2.Producer
	StorageDir  string
}

func NewUploadHandler(repo *repository.ImagesRepo, kw *kafka2.Producer, StorageDir string) *UploadHandler {
	return &UploadHandler{
		Repo:        repo,
		KafkaWriter: kw,
		StorageDir:  StorageDir,
	}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println("Upload: failed to close file:", err)
		}
	}(file)

	dirs := []string{
		filepath.Join(h.StorageDir, "original"),
		filepath.Join(h.StorageDir, "processed"),
		filepath.Join(h.StorageDir, "thumbs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			http.Error(w, "failed to create directory", http.StatusInternalServerError)
			return
		}
	}

	storedPath := filepath.Join(h.StorageDir, "original", header.Filename)
	out, err := os.Create(storedPath)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Println("Upload: failed to close file:", err)
		}
	}(out)

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	id, err := h.Repo.Create(storedPath, header.Filename, "jpeg")
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	msg := kafka.Message{
		Value: []byte(strconv.Itoa(int(id))),
	}
	if err := h.KafkaWriter.Write(msg); err != nil {
		http.Error(w, "Kafka error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%d", id)))
	if err != nil {
		log.Printf("Upload: failed to write response: %s", err)
	}
}

func (h *UploadHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	image, err := h.Repo.GetByID(int64(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, *image.ProcessedPath)
}

func (h *UploadHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	image, err := h.Repo.GetByID(int64(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(*image.StoredPath); err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	if err := h.Repo.Delete(int64(id)); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
