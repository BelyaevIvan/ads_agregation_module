package handler

import (
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/service"
)

type FiltersHandler struct {
	filters *service.FiltersService
}

func NewFiltersHandler(filters *service.FiltersService) *FiltersHandler {
	return &FiltersHandler{filters: filters}
}

// Sizes handles GET /api/v1/filters/sizes
func (h *FiltersHandler) Sizes(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	sizes, err := h.filters.GetSizes()
	if err != nil {
		return err
	}
	return middleware.JSON(w, http.StatusOK, sizes)
}
