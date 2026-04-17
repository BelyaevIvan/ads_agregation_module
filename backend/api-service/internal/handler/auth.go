package handler

import (
	"encoding/json"
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	user, err := h.auth.Register(&req)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusCreated, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return middleware.NewAppError(http.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.auth.Login(&req)
	if err != nil {
		return err
	}

	return middleware.JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return middleware.NewAppError(http.StatusMethodNotAllowed, "метод не поддерживается")
	}

	return middleware.JSON(w, http.StatusOK, map[string]string{
		"message": "выход выполнен успешно",
	})
}
