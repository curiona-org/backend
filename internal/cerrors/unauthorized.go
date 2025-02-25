package cerrors

import "net/http"

var Unauthorized = &AppError{
	code:    http.StatusUnauthorized,
	message: "Unauthorized",
}
