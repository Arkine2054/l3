package processor

import (
	"context"

	"gitlab.com/arkine/l3/4/internal/repository"
)

type MessageConsumer interface {
	Read(ctx context.Context) (string, error)
	Close() error
}

type ImageRepository interface {
	GetByID(id int64) (*repository.Image, error)
	UpdatePathsAndStatus(id int64, p, t *string, status string) error
}
