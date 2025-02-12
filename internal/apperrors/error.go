package apperrors

import "net/http"

type AppError struct {
	code    int
	message string
}

func (e AppError) Code() int {
	return e.code
}

func (e AppError) Error() string {
	return e.message
}

func Wrap(appError error, err error) error {
	var ae *AppError
	if e, ok := appError.(*AppError); ok {
		ae = e
	} else {
		ae = &AppError{code: http.StatusInternalServerError, message: appError.Error()}
	}

	if err == nil {
		return ae
	}

	if ae.message == "" {
		ae.message = err.Error()
	} else {
		ae.message = ae.message + ": " + err.Error()
	}

	return ae
}
