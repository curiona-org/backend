package cerrors

import "net/http"

var NotFound = &AppError{
	code:    http.StatusNotFound,
	message: "Not found",
}
