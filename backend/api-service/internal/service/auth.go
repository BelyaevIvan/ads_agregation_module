package service

import (
	"net/http"
	"regexp"
	"time"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthService struct {
	users     *repository.UserRepo
	jwtSecret string
}

func NewAuthService(users *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(req *model.RegisterRequest) (*model.User, error) {
	if req.Email == "" {
		return nil, middleware.NewAppError(http.StatusBadRequest, "email обязателен")
	}
	if !emailRegex.MatchString(req.Email) {
		return nil, middleware.NewAppError(http.StatusBadRequest, "некорректный формат email")
	}
	if len(req.Password) < 8 {
		return nil, middleware.NewAppError(http.StatusBadRequest, "пароль должен содержать минимум 8 символов")
	}

	existing, err := s.users.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, middleware.NewAppError(http.StatusConflict, "пользователь с таким email уже существует")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	return s.users.Create(req.Email, string(hash))
}

func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, middleware.NewAppError(http.StatusBadRequest, "email и пароль обязательны")
	}

	user, err := s.users.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, middleware.NewAppError(http.StatusUnauthorized, "неверный email или пароль")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, middleware.NewAppError(http.StatusUnauthorized, "неверный email или пароль")
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expiresAt.Unix(),
	})
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}
