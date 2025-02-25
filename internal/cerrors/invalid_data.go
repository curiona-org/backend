package cerrors

import "net/http"

var InvalidData = &AppError{
	code:    http.StatusBadRequest,
	message: "Invalid Data",
}
