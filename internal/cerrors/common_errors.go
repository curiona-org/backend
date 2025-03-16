package cerrors

import "net/http"

var (
	ErrNotFound = &curionaError{
		Code:            "NOT_FOUND",
		StatusCode:      http.StatusNotFound,
		ExternalMessage: "Not found",
	}

	ErrDuplicateData = &curionaError{
		Code:            "DUPLICATE_DATA",
		StatusCode:      http.StatusConflict,
		ExternalMessage: "The data you provided already exists",
	}

	ErrInvalidData = &curionaError{
		Code:            "INVALID_DATA",
		StatusCode:      http.StatusBadRequest,
		ExternalMessage: "Invalid data provided",
	}

	ErrValidation = &curionaError{
		Code:            "VALIDATION_ERROR",
		StatusCode:      http.StatusUnprocessableEntity,
		ExternalMessage: "There was an issue with the data you provided. Please check and try again.",
	}

	ErrInternal = &curionaError{
		Code:            "INTERNAL_ERROR",
		StatusCode:      http.StatusInternalServerError,
		ExternalMessage: "Oops! We encountered an unexpected error. Please try again.",
	}
)
