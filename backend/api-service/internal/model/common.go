package model

import "time"

// PaginatedResponse is the standard paginated list envelope.
type PaginatedResponse struct {
	Items  any `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// APIError is the standard error response.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	URL        string `json:"url"`
	Message    string `json:"message"`
	Date       string `json:"date"`
}

func NewAPIError(code int, url, message string) *APIError {
	return &APIError{
		StatusCode: code,
		URL:        url,
		Message:    message,
		Date:       time.Now().UTC().Format(time.RFC3339),
	}
}
