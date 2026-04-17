package service

import (
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"
)

type StatsService struct {
	stats *repository.StatsRepo
}

func NewStatsService(stats *repository.StatsRepo) *StatsService {
	return &StatsService{stats: stats}
}

func (s *StatsService) GetStats() (*model.Stats, error) {
	return s.stats.GetStats()
}

func (s *StatsService) ListingsByDay(days int) ([]model.ListingsByDay, error) {
	validDays := map[int]bool{7: true, 30: true, 90: true}
	if !validDays[days] {
		return nil, middleware.NewAppError(http.StatusBadRequest, "параметр 'days' может быть 7, 30 или 90")
	}
	return s.stats.ListingsByDay(days)
}

func (s *StatsService) TopBrands(limit int) ([]model.BrandCount, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' не может превышать 100")
	}
	return s.stats.TopBrands(limit)
}

func (s *StatsService) TopCities(limit int) ([]model.CityCount, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' не может превышать 100")
	}
	return s.stats.TopCities(limit)
}
