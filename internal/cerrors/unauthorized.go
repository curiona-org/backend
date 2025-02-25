package cerrors

import "net/http"

var ErrUnauthorized = &AppError{
	ErrorCode:       http.StatusUnauthorized,
	ExternalMessage: "Unauthorized",
}
