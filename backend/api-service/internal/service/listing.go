package service

import (
	"net/http"
	"regexp"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

type ListingService struct {
	listings *repository.ListingRepo
}

func NewListingService(listings *repository.ListingRepo) *ListingService {
	return &ListingService{listings: listings}
}

func (s *ListingService) Search(p *model.ListingSearchParams) ([]model.ListingListItem, int, error) {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' не может превышать 100")
	}
	if p.Offset < 0 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть неотрицательным")
	}
	if p.Sort == "" {
		p.Sort = "date_desc"
	}
	validSorts := map[string]bool{"date_desc": true, "price_asc": true, "price_desc": true}
	if !validSorts[p.Sort] {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "недопустимое значение параметра 'sort'")
	}
	if p.Condition != "" && p.Condition != "new" && p.Condition != "used" {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'condition' может быть 'new' или 'used'")
	}

	return s.listings.Search(p)
}

func (s *ListingService) GetByID(id string) (*model.ListingDetail, error) {
	if !IsValidUUID(id) {
		return nil, middleware.NewAppError(http.StatusBadRequest, "некорректный формат UUID")
	}
	detail, err := s.listings.GetByID(id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, middleware.NewAppError(http.StatusNotFound, "объявление не найдено")
	}
	return detail, nil
}

func (s *ListingService) SetVisibility(id string, isHidden *bool) error {
	if !IsValidUUID(id) {
		return middleware.NewAppError(http.StatusBadRequest, "некорректный формат UUID")
	}
	if isHidden == nil {
		return middleware.NewAppError(http.StatusBadRequest, "поле 'is_hidden' обязательно")
	}

	exists, err := s.listings.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return middleware.NewAppError(http.StatusNotFound, "объявление не найдено")
	}

	return s.listings.SetVisibility(id, *isHidden)
}

func (s *ListingService) UpdateText(id, text string) error {
	if !IsValidUUID(id) {
		return middleware.NewAppError(http.StatusBadRequest, "некорректный формат UUID")
	}
	if text == "" {
		return middleware.NewAppError(http.StatusBadRequest, "поле 'original_text' не может быть пустым")
	}

	exists, err := s.listings.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return middleware.NewAppError(http.StatusNotFound, "объявление не найдено")
	}

	return s.listings.UpdateText(id, text)
}

func (s *ListingService) AdminSearch(p *model.AdminListingSearchParams) ([]model.AdminListingItem, int, error) {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' не может превышать 100")
	}
	if p.Offset < 0 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть неотрицательным")
	}
	if p.Sort == "" {
		p.Sort = "date_desc"
	}
	validSorts := map[string]bool{"date_desc": true, "price_asc": true, "price_desc": true}
	if !validSorts[p.Sort] {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "недопустимое значение параметра 'sort'")
	}
	if p.Status != "" && p.Status != "active" && p.Status != "hidden" {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'status' может быть 'active', 'hidden' или пустым")
	}

	return s.listings.AdminSearch(p)
}
