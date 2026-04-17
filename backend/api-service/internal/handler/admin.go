package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/service"
)

type AdminHandler struct {
	listings *service.ListingService
	sources  *service.SourceService
	stats    *service.StatsService
	photos   *service.PhotoService
}

func NewAdminHandler(
	listings *service.ListingService,
	sources *service.SourceService,
	stats *service.StatsService,
	photos *service.PhotoService,
) *AdminHandler {
	return &AdminHandler{listings: listings, sources: sources, stats: stats, photos: photos}
}

// AdminListings handles GET /api/v1/admin/listings
func (h *AdminHandler) AdminListings(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	q := r.URL.Query()
	p := &model.AdminListingSearchParams{
		Q:         q.Get("q"),
		Status:    q.Get("status"),
		Platforms: q["platform"],
		Sort:      q.Get("sort"),
	}

	var err error
	p.Limit, err = parseInt(q.Get("limit"), 20)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' должен быть числом")
	}
	p.Offset, err = parseInt(q.Get("offset"), 0)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть числом")
	}

	items, total, err := h.listings.AdminSearch(p)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, model.PaginatedResponse{
		Items:  items,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	})
}

// Visibility handles PATCH /api/v1/admin/listings/{id}/visibility
func (h *AdminHandler) Visibility(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	// /api/v1/admin/listings/{id}/visibility
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/listings/")
	id := strings.TrimSuffix(path, "/visibility")

	var req model.VisibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.listings.SetVisibility(id, req.IsHidden); err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, map[string]any{
		"id":        id,
		"is_hidden": *req.IsHidden,
	})
}

// EditText handles PATCH /api/v1/admin/listings/{id}/text
func (h *AdminHandler) EditText(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/listings/")
	id := strings.TrimSuffix(path, "/text")

	var req model.EditTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.listings.UpdateText(id, req.OriginalText); err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, map[string]any{
		"id":            id,
		"original_text": req.OriginalText,
	})
}

// DeletePhoto handles DELETE /api/v1/admin/listings/{id}/photos/{photo_id}
func (h *AdminHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	// /api/v1/admin/listings/{id}/photos/{photo_id}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/listings/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "photos" {
		return middleware.NewAppError(http.StatusBadRequest, "некорректный путь запроса")
	}
	listingID := parts[0]
	photoID := parts[2]

	newCoverID, err := h.photos.DeletePhoto(listingID, photoID)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, map[string]any{
		"deleted_photo_id": photoID,
		"new_cover_id":     newCoverID,
	})
}

// Stats handles GET /api/v1/admin/stats
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	s, err := h.stats.GetStats()
	if err != nil {
		return err
	}
	return middleware.JSON(w, http.StatusOK, s)
}

// ListingsByDay handles GET /api/v1/admin/stats/listings-by-day
func (h *AdminHandler) ListingsByDay(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	days, err := parseInt(r.URL.Query().Get("days"), 30)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'days' должен быть числом")
	}

	items, err := h.stats.ListingsByDay(days)
	if err != nil {
		return err
	}
	return middleware.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// TopBrands handles GET /api/v1/admin/stats/top-brands
func (h *AdminHandler) TopBrands(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	limit, err := parseInt(r.URL.Query().Get("limit"), 10)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' должен быть числом")
	}

	items, err := h.stats.TopBrands(limit)
	if err != nil {
		return err
	}
	return middleware.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// TopCities handles GET /api/v1/admin/stats/top-cities
func (h *AdminHandler) TopCities(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	limit, err := parseInt(r.URL.Query().Get("limit"), 10)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' должен быть числом")
	}

	items, err := h.stats.TopCities(limit)
	if err != nil {
		return err
	}
	return middleware.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// Sources handles GET/POST /api/v1/admin/sources
func (h *AdminHandler) Sources(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return h.listSources(w, r)
	case http.MethodPost:
		return h.createSource(w, r)
	default:
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}

func (h *AdminHandler) listSources(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	limit, err := parseInt(q.Get("limit"), 50)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' должен быть числом")
	}
	offset, err := parseInt(q.Get("offset"), 0)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть числом")
	}

	items, total, err := h.sources.List(limit, offset)
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

func (h *AdminHandler) createSource(w http.ResponseWriter, r *http.Request) error {
	var req model.CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	src, err := h.sources.Create(&req)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusCreated, src)
}

// ToggleSource handles PATCH /api/v1/admin/sources/{id}/toggle
func (h *AdminHandler) ToggleSource(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/sources/")
	id := strings.TrimSuffix(path, "/toggle")

	var req model.ToggleSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.sources.Toggle(id, req.IsActive); err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, map[string]any{
		"id":        id,
		"is_active": *req.IsActive,
	})
}
