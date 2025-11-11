package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"

	"gitlab.com/arkine/l3/4/internal/handlers"
	"gitlab.com/arkine/l3/4/internal/processor"
	"gitlab.com/arkine/l3/4/internal/repository"
)

func main() {
	ctx := context.Background()

	// Подключение к базе и миграции внутри репозитория
	repo, err := repository.NewImagesRepo()
	if err != nil {
		log.Fatal("❌ DB init failed:", err)
	}
	defer repo.DB.Close()

	// Kafka Writer (Producer)
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKER")}
	if kafkaBrokers[0] == "" {
		kafkaBrokers = []string{"kafka:9092"}
	}
	kWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  kafkaBrokers,
		Topic:    "images",
		Balancer: &kafka.LeastBytes{},
	})
	defer kWriter.Close()

	// Kafka Reader (Consumer) — наш фоновой воркер
	kReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kafkaBrokers,
		Topic:    "images",
		GroupID:  "image-workers-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer kReader.Close()

	storageDir := "./data"

	// Запуск фонового воркера обработки изображений
	worker := &processor.Worker{
		Repo:          repo,
		Reader:        kReader,
		StorageDir:    storageDir,
		WatermarkText: os.Getenv("WATERMARK_TEXT"),
	}
	worker.Start(ctx)

	// Обработчик загрузки
	uploadHandler := &handlers.UploadHandler{
		Repo:        repo,
		KafkaWriter: kWriter,
		StorageDir:  storageDir,
	}

	// HTTP роутер
	r := mux.NewRouter()

	// Статика для web интерфейса
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Статика для изображений
	r.PathPrefix("/data/processed/").Handler(
		http.StripPrefix("/data/processed/", http.FileServer(http.Dir("./data/processed"))),
	)
	r.PathPrefix("/data/thumbs/").Handler(
		http.StripPrefix("/data/thumbs/", http.FileServer(http.Dir("./data/thumbs"))),
	)
	r.PathPrefix("/data/original/").Handler(
		http.StripPrefix("/data/original/", http.FileServer(http.Dir("./data/original"))),
	)

	// Главная страница
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	// API
	r.HandleFunc("/upload", uploadHandler.Upload).Methods("POST")
	r.HandleFunc("/image/{id:[0-9]+}", uploadHandler.GetImage).Methods("GET")
	r.HandleFunc("/image/{id:[0-9]+}", uploadHandler.DeleteImage).Methods("DELETE")

	// Список изображений
	r.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		images, err := repo.List()
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(images)
	}).Methods("GET")

	// HTTP сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srvAddr := ":" + port

	srv := &http.Server{
		Addr:         srvAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("🚀 Server running on", srvAddr)
	log.Fatal(srv.ListenAndServe())
}
