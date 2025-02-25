package cerrors

import "net/http"

var ErrInvalidData = &AppError{
	code:    http.StatusBadRequest,
	message: "Invalid Data",
}
