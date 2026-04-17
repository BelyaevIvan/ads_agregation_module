package handler

import (
	"encoding/json"
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	userID := middleware.UserIDFromContext(r.Context())
	user, err := h.users.GetProfile(userID)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPut {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	userID := middleware.UserIDFromContext(r.Context())

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	user, err := h.users.UpdateProfile(userID, &req)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, user)
}

// Me handles both GET and PUT on /api/v1/users/me
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return h.GetProfile(w, r)
	case http.MethodPut:
		return h.UpdateProfile(w, r)
	default:
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}
