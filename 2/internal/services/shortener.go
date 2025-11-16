package services

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/arkine/l3/2/internal/repo"
	"gitlab.com/arkine/l3/2/internal/utils"
)

type ShortenerService struct {
	repo *repo.ShortURLRepo
}

func NewShortenerService(r *repo.ShortURLRepo) *ShortenerService {
	return &ShortenerService{repo: r}
}

type CreateResponse struct {
	Alias string `json:"alias"`
	URL   string `json:"url"`
}

func (s *ShortenerService) CreateShortURL(target, customAlias string, r *http.Request) (*CreateResponse, error) {
	if target == "" {
		return nil, errors.New("target required")
	}

	alias := customAlias
	if alias == "" {
		alias = utils.RandomAlias(6)
	}

	existing, _ := s.repo.FindByAlias(alias)
	if existing != nil {
		return nil, fmt.Errorf("alias '%s' already exists", alias)
	}

	_, err := s.repo.Create(alias, target)
	if err != nil {
		return nil, err
	}

	return &CreateResponse{
		Alias: alias,
		URL:   fmt.Sprintf("%s/s/%s", utils.ServerBaseURL(r), alias),
	}, nil
}
