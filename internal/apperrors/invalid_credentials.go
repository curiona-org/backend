package apperrors

import (
	"net/http"
)

func InvalidCredentials() *AppError {
	return &AppError{
		code:    http.StatusUnauthorized,
		message: "Invalid Credentials",
	}
}
