package apperrors

import "net/http"

func Unauthorized() *AppError {
	return &AppError{
		code:    http.StatusUnauthorized,
		message: "Unauthorized",
	}
}
