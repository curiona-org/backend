package cerrors

import "net/http"

var (
	ErrInvalidCredentials = &CurionaError{
		ErrorCode:       http.StatusUnauthorized,
		ExternalMessage: "Invalid Credentials",
	}

	ErrUnauthorized = &CurionaError{
		ErrorCode:       http.StatusUnauthorized,
		ExternalMessage: "Unauthorized",
	}
)
