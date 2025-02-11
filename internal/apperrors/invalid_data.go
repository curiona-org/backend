package apperrors

import "net/http"

func InvalidData() error {
	return &AppError{
		code:    http.StatusBadRequest,
		message: "Invalid Data",
	}
}
