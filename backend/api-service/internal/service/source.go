package service

import (
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"
)

type SourceService struct {
	sources *repository.SourceRepo
}

func NewSourceService(sources *repository.SourceRepo) *SourceService {
	return &SourceService{sources: sources}
}

func (s *SourceService) List(limit, offset int) ([]model.Source, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' не может превышать 100")
	}
	if offset < 0 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть неотрицательным")
	}
	return s.sources.List(limit, offset)
}

func (s *SourceService) Create(req *model.CreateSourceRequest) (*model.Source, error) {
	if req.Platform != "telegram" && req.Platform != "vk" {
		return nil, middleware.NewAppError(http.StatusBadRequest, "platform должен быть 'telegram' или 'vk'")
	}
	if req.ExternalID == "" {
		return nil, middleware.NewAppError(http.StatusBadRequest, "external_id обязателен")
	}

	exists, err := s.sources.ExistsByPlatformAndExtID(req.Platform, req.ExternalID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, middleware.NewAppError(http.StatusConflict, "источник с такой комбинацией platform/external_id уже существует")
	}

	return s.sources.Create(req.Platform, req.ExternalID, req.Title)
}

func (s *SourceService) Toggle(id string, isActive *bool) error {
	if !IsValidUUID(id) {
		return middleware.NewAppError(http.StatusBadRequest, "некорректный формат UUID")
	}
	if isActive == nil {
		return middleware.NewAppError(http.StatusBadRequest, "поле 'is_active' обязательно")
	}
	exists, err := s.sources.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return middleware.NewAppError(http.StatusNotFound, "источник не найден")
	}
	return s.sources.SetActive(id, *isActive)
}
