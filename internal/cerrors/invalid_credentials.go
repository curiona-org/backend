package cerrors

import (
	"net/http"
)

var InvalidCredentials = &AppError{
	code:    http.StatusUnauthorized,
	message: "Invalid Credentials",
}
