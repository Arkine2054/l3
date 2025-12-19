package main

import (
	"context"
	"log"
	"os"
	"time"

	"gitlab.com/arkine/l3/4/internal/kafka"
	"gitlab.com/arkine/l3/4/internal/processor"
	"gitlab.com/arkine/l3/4/internal/repository"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var repo *repository.ImagesRepo
	var err error
	for i := 0; i < 10; i++ {
		repo, err = repository.NewImagesRepo()
		if err == nil {
			log.Println("Worker: connected to DB")
			break
		}
		log.Println("Worker: DB not ready, retrying in 2s:", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Worker: failed to connect to DB after retries: %v", err)
	}

	kafkaBrokers := []string{os.Getenv("KAFKA_BROKER")}
	var consumer *kafka.Consumer
	for i := 0; i < 10; i++ {
		consumer = kafka.NewConsumer(kafkaBrokers, "images", "image-workers")
		if consumer != nil {
			log.Println("Worker: Kafka consumer created")
			break
		}
		log.Println("Worker: Kafka not ready, retrying in 2s...")
		time.Sleep(2 * time.Second)
	}
	if consumer == nil {
		log.Fatal("Worker: failed to create Kafka consumer")
	}

	worker := &processor.Worker{
		Repo:          repo,
		Consumer:      consumer,
		StorageDir:    os.Getenv("STORAGE_DIR"),
		WatermarkText: os.Getenv("WATERMARK_TEXT"),
		FontPath:      "/app/internal/assets/font.ttf",
	}

	worker.Run(ctx)
}
