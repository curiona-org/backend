package cerrors

import "net/http"

var ErrNotFound = &AppError{
	ErrorCode:       http.StatusNotFound,
	ExternalMessage: "Not found",
}
