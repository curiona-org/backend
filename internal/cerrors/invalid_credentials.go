package cerrors

import (
	"net/http"
)

var ErrInvalidCredentials = &AppError{
	ErrorCode:       http.StatusUnauthorized,
	ExternalMessage: "Invalid Credentials",
}
