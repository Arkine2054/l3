package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gitlab.com/arkine/l3/4/internal/handlers"
	"gitlab.com/arkine/l3/4/internal/kafka"
	"gitlab.com/arkine/l3/4/internal/repository"
	"gitlab.com/arkine/l3/4/internal/router"
)

func main() {

	var repo *repository.ImagesRepo
	var err error
	for i := 0; i < 10; i++ {
		repo, err = repository.NewImagesRepo()
		if err == nil {
			log.Println("API: connected to DB")
			break
		}
		log.Println("API: DB not ready, retrying in 2s:", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("API: failed to connect to DB after retries: %v", err)
	}

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = "/data"
	}

	kWriter := kafka.NewProducer(
		[]string{os.Getenv("KAFKA_BROKER")},
		"images",
	)
	defer func(kWriter *kafka.Producer) {
		err := kWriter.Close()
		if err != nil {
			log.Printf("failed to close kafka writer: %v", err)
		}
	}(kWriter)

	uploadHandler := handlers.NewUploadHandler(
		repo,
		kWriter,
		storageDir,
	)

	r := router.NewRouter(uploadHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("API server running on :" + port)
	log.Fatal(srv.ListenAndServe())
}
