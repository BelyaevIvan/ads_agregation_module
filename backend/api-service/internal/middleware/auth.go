package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ContextUserID contextKey = "user_id"
	ContextRole   contextKey = "role"
)

// Auth validates JWT and injects user_id + role into the request context.
func Auth(secret string) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			header := r.Header.Get("Authorization")
			if header == "" {
				return NewAppError(http.StatusUnauthorized, "отсутствует токен авторизации")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return NewAppError(http.StatusUnauthorized, "неверный формат заголовка Authorization")
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, NewAppError(http.StatusUnauthorized, "неподдерживаемый алгоритм подписи токена")
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return NewAppError(http.StatusUnauthorized, "невалидный или просроченный токен")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return NewAppError(http.StatusUnauthorized, "невалидный токен")
			}

			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string)
			if userID == "" {
				return NewAppError(http.StatusUnauthorized, "невалидный токен: отсутствует user_id")
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextRole, role)
			return next(w, r.WithContext(ctx))
		}
	}
}

// RequireAdmin checks that the authenticated user has admin role.
func RequireAdmin(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		role, _ := r.Context().Value(ContextRole).(string)
		if role != "admin" {
			return NewAppError(http.StatusForbidden, "доступ запрещён: требуется роль администратора")
		}
		return next(w, r)
	}
}

// UserIDFromContext extracts user_id from context.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextUserID).(string)
	return v
}
