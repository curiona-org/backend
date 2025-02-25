package cerrors

import "net/http"

var ErrUnauthorized = &AppError{
	code:    http.StatusUnauthorized,
	message: "Unauthorized",
}
