package cerrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	DefaultErrorCode    = http.StatusInternalServerError
	DefaultErrorMessage = "Oops! We encountered an unexpected error. Please try again."
)

type CurionaError struct {
	// ErrorCode is the HTTP status code that is returned to the end user.
	ErrorCode int
	// InternalMessage is the message that is logged in the application logs.
	// It should not be exposed to the end user.
	InternalMessage string
	// ExternalMessage is the error message that is returned to the end user.
	ExternalMessage string
}

// New creates a new application error with the provided error string.
func New(errString string) error {
	return &CurionaError{
		ErrorCode:       DefaultErrorCode,
		InternalMessage: errString,
		ExternalMessage: DefaultErrorMessage,
	}
}

// Wrap creates a new wrapped application error with the provided error.
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	return &CurionaError{
		ErrorCode:       DefaultErrorCode,
		InternalMessage: err.Error(),
		ExternalMessage: DefaultErrorMessage,
	}
}

// Unwrap converts the application error to the underlying error.
// If the error is not an application error, it returns the original error.
func Unwrap(err error) error {
	var ae *CurionaError
	if errors.As(err, &ae) {
		return ae
	}
	return err
}

func (e *CurionaError) Code() int {
	return e.ErrorCode
}

// Error implements the error interface.
func (e *CurionaError) Error() string {
	return e.InternalMessage
}

func (e *CurionaError) Message() string {
	if e.ExternalMessage == "" {
		return DefaultErrorMessage
	}

	return e.ExternalMessage
}

// With creates a new wrapped application error with the provided error.
func (e *CurionaError) With(err error) error {
	if err == nil {
		return e
	}

	return &CurionaError{
		ErrorCode:       e.ErrorCode,
		InternalMessage: e.InternalMessage + ": " + err.Error(),
		ExternalMessage: e.ExternalMessage,
	}
}

// Msg creates a new wrapped application error with the provided message.
func (e *CurionaError) Msg(errorString string) error {
	return &CurionaError{
		ErrorCode:       e.ErrorCode,
		InternalMessage: e.InternalMessage,
		ExternalMessage: e.ExternalMessage + ": " + errorString,
	}
}

// Msgf creates a new wrapped application error with the provided formatted message.
func (e *CurionaError) Msgf(format string, a ...any) error {
	format = fmt.Sprintf(format, a...)
	return &CurionaError{
		ErrorCode:       e.ErrorCode,
		InternalMessage: e.InternalMessage,
		ExternalMessage: e.ExternalMessage + ": " + format,
	}
}
