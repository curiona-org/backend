package cerrors

import "net/http"

var ErrDuplicateData = &AppError{
	ErrorCode:       http.StatusConflict,
	ExternalMessage: "data already exists",
}
