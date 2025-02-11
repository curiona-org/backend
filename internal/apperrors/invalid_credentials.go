package apperrors

import (
	"net/http"
)

func InvalidCredentials() error {
	return &AppError{
		code:    http.StatusUnauthorized,
		message: "Invalid Credentials",
	}
}
