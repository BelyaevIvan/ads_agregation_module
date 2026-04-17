package service

import (
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"
)

type FavoriteService struct {
	favorites *repository.FavoriteRepo
	listings  *repository.ListingRepo
}

func NewFavoriteService(favorites *repository.FavoriteRepo, listings *repository.ListingRepo) *FavoriteService {
	return &FavoriteService{favorites: favorites, listings: listings}
}

func (s *FavoriteService) List(userID string, limit, offset int) ([]model.FavoriteWithListing, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' не может превышать 100")
	}
	if offset < 0 {
		return nil, 0, middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть неотрицательным")
	}
	return s.favorites.List(userID, limit, offset)
}

func (s *FavoriteService) Add(userID, listingID string) (*model.Favorite, bool, error) {
	if !IsValidUUID(listingID) {
		return nil, false, middleware.NewAppError(http.StatusBadRequest, "некорректный формат listing_id")
	}
	if listingID == "" {
		return nil, false, middleware.NewAppError(http.StatusBadRequest, "listing_id обязателен")
	}

	exists, err := s.listings.Exists(listingID)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, middleware.NewAppError(http.StatusNotFound, "объявление не найдено")
	}

	return s.favorites.Add(userID, listingID)
}

func (s *FavoriteService) Remove(userID, listingID string) error {
	if !IsValidUUID(listingID) {
		return middleware.NewAppError(http.StatusBadRequest, "некорректный формат listing_id")
	}
	removed, err := s.favorites.Remove(userID, listingID)
	if err != nil {
		return err
	}
	if !removed {
		return middleware.NewAppError(http.StatusNotFound, "объявление не найдено в избранном")
	}
	return nil
}
