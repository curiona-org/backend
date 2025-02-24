package cerrors

import (
	"errors"
	"net/http"
)

type AppError struct {
	code    int
	message string
	err     error
}

func (e AppError) Code() int {
	return e.code
}

func (e AppError) Error() string {
	return e.message
}

func Wrap(appError error, err error) error {
	var ae *AppError
	if !errors.As(appError, &ae) {
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

	ae.err = err

	return ae
}

func Unwrap(err error) error {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.err
	}
	return err
}
