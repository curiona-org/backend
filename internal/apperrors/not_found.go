package apperrors

import "net/http"

func NotFound() error {
	return &AppError{
		code:    http.StatusNotFound,
		message: "Not found",
	}
}
