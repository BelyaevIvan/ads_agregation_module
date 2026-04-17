package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/service"
)

type FavoriteHandler struct {
	favorites *service.FavoriteService
}

func NewFavoriteHandler(favorites *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

// List handles GET /api/v1/users/me/favorites
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	userID := middleware.UserIDFromContext(r.Context())
	q := r.URL.Query()
	limit, err := parseInt(q.Get("limit"), 20)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' должен быть числом")
	}
	offset, err := parseInt(q.Get("offset"), 0)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть числом")
	}

	items, total, err := h.favorites.List(userID, limit, offset)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, model.PaginatedResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// Add handles POST /api/v1/users/me/favorites
func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	userID := middleware.UserIDFromContext(r.Context())

	var req model.AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	fav, created, err := h.favorites.Add(userID, req.ListingID)
	if err != nil {
		return err
	}

	code := http.StatusOK
	if created {
		code = http.StatusCreated
	}
	return middleware.JSON(w, code, fav)
}

// Favorites handles both GET and POST on /api/v1/users/me/favorites
func (h *FavoriteHandler) Favorites(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return h.List(w, r)
	case http.MethodPost:
		return h.Add(w, r)
	default:
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}

// Remove handles DELETE /api/v1/users/me/favorites/{listing_id}
func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	userID := middleware.UserIDFromContext(r.Context())
	listingID := strings.TrimPrefix(r.URL.Path, "/api/v1/users/me/favorites/")

	err := h.favorites.Remove(userID, listingID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
