package cerrors

import (
	"net/http"
)

var ErrInvalidCredentials = &AppError{
	code:    http.StatusUnauthorized,
	message: "Invalid Credentials",
}
