package cerrors

import "net/http"

var DuplicateData = &AppError{
	code:    http.StatusConflict,
	message: "data already exists",
}
