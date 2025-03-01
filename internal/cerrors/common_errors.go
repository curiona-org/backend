package cerrors

import "net/http"

var (
	ErrNotFound = &CurionaError{
		ErrorCode:       http.StatusNotFound,
		ExternalMessage: "Not found",
	}

	ErrDuplicateData = &CurionaError{
		ErrorCode:       http.StatusConflict,
		ExternalMessage: "data already exists",
	}

	ErrInvalidData = &CurionaError{
		ErrorCode:       http.StatusBadRequest,
		ExternalMessage: "Invalid Data",
	}
)
