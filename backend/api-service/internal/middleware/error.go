package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// AppError is used by handlers to signal an error with a specific status code.
type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string { return e.Message }

func NewAppError(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// HandlerFunc is an http handler that returns an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorHandler wraps a HandlerFunc and converts returned errors into the
// standard JSON error format described in the spec.
func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		code := http.StatusInternalServerError
		message := "внутренняя ошибка сервера, попробуйте позже"

		if appErr, ok := err.(*AppError); ok {
			code = appErr.Code
			message = appErr.Message
		} else {
			log.Printf("ERROR %s %s: %v", r.Method, r.URL.String(), err)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"statusCode": code,
			"url":        r.URL.String(),
			"message":    message,
			"date":       time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}
