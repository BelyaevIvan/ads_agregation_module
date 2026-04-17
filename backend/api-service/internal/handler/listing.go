package handler

import (
	"net/http"
	"strconv"
	"strings"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/service"
)

type ListingHandler struct {
	listings *service.ListingService
}

func NewListingHandler(listings *service.ListingService) *ListingHandler {
	return &ListingHandler{listings: listings}
}

func (h *ListingHandler) Search(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	q := r.URL.Query()
	p := &model.ListingSearchParams{
		Q:              q.Get("q"),
		Brands:         q["brand"],
		Categories:     q["category"],
		Cities:         q["city"],
		Condition:      q.Get("condition"),
		SizeRus:        q["size_rus"],
		SizeEU:         q["size_eu"],
		SizeUS:         q["size_us"],
		Platforms:      q["platform"],
		Sort:           q.Get("sort"),
		IncludeNoSize:  parseBool(q.Get("include_no_size"), true),
		IncludeNoPrice: parseBool(q.Get("include_no_price"), true),
		IncludeNoCity:  parseBool(q.Get("include_no_city"), true),
	}

	var err error
	p.Limit, err = parseInt(q.Get("limit"), 20)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'limit' должен быть положительным числом")
	}
	p.Offset, err = parseInt(q.Get("offset"), 0)
	if err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "параметр 'offset' должен быть неотрицательным числом")
	}

	if v := q.Get("price_min"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return middleware.NewAppError(http.StatusBadRequest, "параметр 'price_min' должен быть числом")
		}
		p.PriceMin = &f
	}
	if v := q.Get("price_max"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return middleware.NewAppError(http.StatusBadRequest, "параметр 'price_max' должен быть числом")
		}
		p.PriceMax = &f
	}

	items, total, err := h.listings.Search(p)
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

func (h *ListingHandler) GetByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	id := extractPathParam(r.URL.Path, "/api/v1/listings/")
	detail, err := h.listings.GetByID(id)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, detail)
}

func parseBool(s string, defaultVal bool) bool {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseInt(s string, defaultVal int) (int, error) {
	if s == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(s)
}

func extractPathParam(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	return s
}
