package cerrors

import "net/http"

func ResourceNotFound(resource string) *AppError {
	return &AppError{
		code:    http.StatusNotFound,
		message: resource + " not found",
	}
}
