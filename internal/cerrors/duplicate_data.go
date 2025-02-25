package cerrors

import "net/http"

var ErrDuplicateData = &AppError{
	code:    http.StatusConflict,
	message: "data already exists",
}
