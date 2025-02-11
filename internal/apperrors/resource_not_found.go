package apperrors

import "net/http"

func ResourceNotFound(resource string) error {
	return &AppError{
		code:    http.StatusNotFound,
		message: resource + " not found",
	}
}
