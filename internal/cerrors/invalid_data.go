package cerrors

import "net/http"

var ErrInvalidData = &AppError{
	ErrorCode:       http.StatusBadRequest,
	ExternalMessage: "Invalid Data",
}
