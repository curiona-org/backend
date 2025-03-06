package cerrors

import "net/http"

var (
	ErrNotFound = &curionaError{
		ErrorCode:       http.StatusNotFound,
		ExternalMessage: "Not found",
	}

	ErrDuplicateData = &curionaError{
		ErrorCode:       http.StatusConflict,
		ExternalMessage: "data already exists",
	}

	ErrInvalidData = &curionaError{
		ErrorCode:       http.StatusBadRequest,
		ExternalMessage: "Invalid Data",
	}
)
