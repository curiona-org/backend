package cerrors_test

import (
	"errors"
	"testing"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/stretchr/testify/assert"
)

func TestCerror_New(t *testing.T) {
	errString := "internal error"
	err := cerrors.New(errString)

	appErr, ok := err.(*cerrors.CurionaError)
	assert.True(t, ok)
	assert.Equal(t, cerrors.DefaultErrorCode, appErr.Code())
	assert.Equal(t, errString, appErr.Error())
	assert.Equal(t, cerrors.DefaultErrorMessage, appErr.Message())
}

func TestCerror_NewEmptyMessage(t *testing.T) {
	errString := "internal error"
	err := cerrors.New(errString)

	appErr, ok := err.(*cerrors.CurionaError)
	assert.True(t, ok)
	appErr.ExternalMessage = ""
	assert.Equal(t, cerrors.DefaultErrorCode, appErr.Code())
	assert.Equal(t, errString, appErr.Error())
	assert.Equal(t, cerrors.DefaultErrorMessage, appErr.Message())
}

func TestCerror_Wrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := cerrors.Wrap(originalErr)

	appErr, ok := err.(*cerrors.CurionaError)
	assert.True(t, ok)
	assert.Equal(t, cerrors.DefaultErrorCode, appErr.Code())
	assert.Equal(t, originalErr.Error(), appErr.Error())
	assert.Equal(t, cerrors.DefaultErrorMessage, appErr.Message())
}

func TestCerror_WrapNilError(t *testing.T) {
	err := cerrors.Wrap(nil)
	assert.NoError(t, err)
}

func TestCerror_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := cerrors.Wrap(originalErr)

	unwrappedErr := cerrors.Unwrap(err)
	assert.Equal(t, err, unwrappedErr)

	nonAppErr := errors.New("non app error")
	unwrappedErr = cerrors.Unwrap(nonAppErr)
	assert.Equal(t, nonAppErr, unwrappedErr)
}

func TestCerror_With(t *testing.T) {
	appErr := &cerrors.CurionaError{
		ErrorCode:       cerrors.DefaultErrorCode,
		InternalMessage: "internal error",
		ExternalMessage: cerrors.DefaultErrorMessage,
	}

	wrappedErr := appErr.With(errors.New("additional error"))
	wrappedAppErr, ok := wrappedErr.(*cerrors.CurionaError)
	assert.True(t, ok)
	assert.Equal(t, appErr.ErrorCode, wrappedAppErr.Code())
	assert.Equal(t, "internal error: additional error", wrappedAppErr.Error())
	assert.Equal(t, appErr.Message(), wrappedAppErr.Message())
}

func TestCerror_WithNil(t *testing.T) {
	appErr := &cerrors.CurionaError{
		ErrorCode:       cerrors.DefaultErrorCode,
		InternalMessage: "internal error",
		ExternalMessage: cerrors.DefaultErrorMessage,
	}

	wrappedErr := appErr.With(nil)
	wrappedAppErr, ok := wrappedErr.(*cerrors.CurionaError)
	assert.True(t, ok)
	assert.Equal(t, appErr.ErrorCode, wrappedAppErr.Code())
	assert.Equal(t, appErr.Error(), wrappedAppErr.Error())
	assert.Equal(t, appErr.Message(), wrappedAppErr.Message())
}

func TestCerror_Msg(t *testing.T) {
	appErr := &cerrors.CurionaError{
		ErrorCode:       cerrors.DefaultErrorCode,
		InternalMessage: "internal error",
		ExternalMessage: cerrors.DefaultErrorMessage,
	}

	wrappedErr := appErr.Msg("additional message")
	wrappedAppErr, ok := wrappedErr.(*cerrors.CurionaError)
	assert.True(t, ok)
	assert.Equal(t, appErr.ErrorCode, wrappedAppErr.Code())
	assert.Equal(t, appErr.Error(), wrappedAppErr.Error())
	assert.Equal(t, appErr.Message()+": additional message", wrappedAppErr.Message())
}

func TestCerror_Msgf(t *testing.T) {
	appErr := &cerrors.CurionaError{
		ErrorCode:       cerrors.DefaultErrorCode,
		InternalMessage: "internal error",
		ExternalMessage: cerrors.DefaultErrorMessage,
	}

	wrappedErr := appErr.Msgf("additional message %d", 123)
	wrappedAppErr, ok := wrappedErr.(*cerrors.CurionaError)
	assert.True(t, ok)
	assert.Equal(t, appErr.ErrorCode, wrappedAppErr.Code())
	assert.Equal(t, appErr.Error(), wrappedAppErr.Error())
	assert.Equal(t, appErr.Message()+": additional message 123", wrappedAppErr.Message())
}
