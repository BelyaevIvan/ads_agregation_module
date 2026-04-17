package service

import (
	"net/http"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"
)

type UserService struct {
	users *repository.UserRepo
}

func NewUserService(users *repository.UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetProfile(userID string) (*model.User, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, middleware.NewAppError(http.StatusNotFound, "пользователь не найден")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(userID string, req *model.UpdateProfileRequest) (*model.User, error) {
	if req.FullName != nil && len(*req.FullName) > 255 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "full_name не может превышать 255 символов")
	}
	if req.Phone != nil && len(*req.Phone) > 30 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "phone не может превышать 30 символов")
	}
	if req.TgLink != nil && len(*req.TgLink) > 255 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "tg_link не может превышать 255 символов")
	}
	if req.VkLink != nil && len(*req.VkLink) > 255 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "vk_link не может превышать 255 символов")
	}

	user, err := s.users.UpdateProfile(userID, req)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, middleware.NewAppError(http.StatusNotFound, "пользователь не найден")
	}
	return user, nil
}
