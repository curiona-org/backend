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

type CurionaError interface {
	error

	// Message returns the error message for the end user.
	Message() string
	// Code returns the HTTP status code of the error.
	Code() int
	// With concats an error to the external message of the error.
	With(err error) error
	// Msg concats an error message to the external message of the error.
	Msg(errorString string) error
	// Msgf concats a formatted error message to the external message of the error.
	Msgf(format string, args ...any) error

	// SetCode overrides the error code of the error.
	SetCode(code int) CurionaError
	// SetInternalMessage overrides the internal message of the error.
	SetInternalMessage(message string) CurionaError
	// SetExternalMessage overrides the external message of the error.
	SetExternalMessage(message string) CurionaError
}

type curionaError struct {
	err error

	// ErrorCode is the HTTP status code that is returned to the end user.
	ErrorCode int
	// InternalMessage should be used for logging and debugging purposes.
	InternalMessage string
	// ExternalMessage is the error message that is returned to the end user.
	ExternalMessage string
}

// New creates a new application error with the provided error.
func New(err any) CurionaError {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case error:
		return &curionaError{
			err:             err.(error),
			ErrorCode:       DefaultErrorCode,
			InternalMessage: err.(error).Error(),
			ExternalMessage: DefaultErrorMessage,
		}
	case string:
		return &curionaError{
			ErrorCode:       DefaultErrorCode,
			InternalMessage: err.(string),
			ExternalMessage: DefaultErrorMessage,
		}
	default:
		return &curionaError{
			ErrorCode:       DefaultErrorCode,
			InternalMessage: fmt.Sprintf("%v", err),
			ExternalMessage: DefaultErrorMessage,
		}
	}
}

// Unwrap returns the original from a CurionaError.
// If the provided error is not a CurionaError, it is returned as is.
func Unwrap(err error) error {
	var cerr *curionaError
	if !errors.As(err, &cerr) {
		return err
	}

	return cerr.err
}

func (e *curionaError) Code() int {
	return e.ErrorCode
}

func (e *curionaError) Error() string {
	return e.InternalMessage
}

func (e *curionaError) Message() string {
	return e.ExternalMessage
}

func (e *curionaError) With(err error) error {
	if err == nil {
		return e
	}

	return &curionaError{
		ErrorCode:       e.ErrorCode,
		InternalMessage: e.InternalMessage + ": " + err.Error(),
		ExternalMessage: e.ExternalMessage + ": " + err.Error(),
		err:             err,
	}
}

// Msg creates a new wrapped application error with the provided message.
func (e *curionaError) Msg(errorString string) error {
	return &curionaError{
		ErrorCode:       e.ErrorCode,
		InternalMessage: e.InternalMessage,
		ExternalMessage: e.ExternalMessage + ": " + errorString,
	}
}

// Msgf creates a new wrapped application error with the provided formatted message.
func (e *curionaError) Msgf(format string, a ...any) error {
	format = fmt.Sprintf(format, a...)
	return &curionaError{
		ErrorCode:       e.ErrorCode,
		InternalMessage: e.InternalMessage,
		ExternalMessage: e.ExternalMessage + ": " + format,
	}
}

func (e *curionaError) SetCode(code int) CurionaError {
	e.ErrorCode = code
	return e
}

func (e *curionaError) SetInternalMessage(message string) CurionaError {
	e.InternalMessage = message
	return e
}

func (e *curionaError) SetExternalMessage(message string) CurionaError {
	e.ExternalMessage = message
	return e
}
