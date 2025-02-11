package apperrors

import "net/http"

func DuplicateData(label string) *AppError {
	return &AppError{
		code:    http.StatusConflict,
		message: label + " already exists",
	}
}
