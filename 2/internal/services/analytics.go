package services

import (
	"gitlab.com/arkine/l3/2/internal/models"
	"gitlab.com/arkine/l3/2/internal/repo"
)

type AnalyticsService struct {
	clickRepo *repo.ClickRepo
}

func NewAnalyticsService(clickRepo *repo.ClickRepo) *AnalyticsService {
	return &AnalyticsService{clickRepo: clickRepo}
}

func (s *AnalyticsService) GetAnalytics(alias string) ([]models.Click, error) {
	return s.clickRepo.GetClicks(alias)
}
