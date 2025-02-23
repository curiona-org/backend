package cerrors

import "net/http"

func InvalidData() *AppError {
	return &AppError{
		code:    http.StatusBadRequest,
		message: "Invalid Data",
	}
}
