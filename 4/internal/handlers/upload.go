package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"

	"gitlab.com/arkine/l3/4/internal/repository"
)

type UploadHandler struct {
	Repo        *repository.ImagesRepo
	KafkaWriter *kafka.Writer
	StorageDir  string
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := header.Filename
	path := filepath.Join(h.StorageDir, filename)

	out, err := os.Create(path)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(filename))
	format := strings.TrimPrefix(ext, ".")

	id, err := h.Repo.Create(path, filename, format)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	msg := kafka.Message{
		Value: []byte(strconv.Itoa(int(id))),
	}
	if err := h.KafkaWriter.WriteMessages(r.Context(), msg); err != nil {
		http.Error(w, "kafka write error", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"id":       id,
		"filename": filename,
		"status":   "uploaded",
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("failed to write response: %v", err)
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
