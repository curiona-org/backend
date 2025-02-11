package apperrors

import "net/http"

func DuplicateData(label string) error {
	return &AppError{
		code:    http.StatusConflict,
		message: label + " already exists",
	}
}
