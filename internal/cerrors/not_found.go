package cerrors

import "net/http"

var ErrNotFound = &AppError{
	code:    http.StatusNotFound,
	message: "Not found",
}
