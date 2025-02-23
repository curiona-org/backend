package cerrors

import "net/http"

func NotFound() *AppError {
	return &AppError{
		code:    http.StatusNotFound,
		message: "Not found",
	}
}
