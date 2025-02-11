package apperrors

import "net/http"

func Unauthorized() error {
	return &AppError{
		code:    http.StatusUnauthorized,
		message: "Unauthorized",
	}
}
