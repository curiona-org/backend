package api

import (
	"errors"
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

//nolint:staticcheck,errorlint,wastedassign
func (a *API) ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *cerrors.AppError
	var httpErr *echo.HTTPError
	code := http.StatusInternalServerError
	switch {
	case errors.As(err, &appErr):
		code = appErr.Code()
	case errors.As(err, &httpErr):
		if httpErr.Code == http.StatusNotFound {
			err = cerrors.NotFound()
		}
		code = httpErr.Code
	}

	var validationErrMsgs []validationErrMsg
	if validationErrs, isValidationErr := err.(validator.ValidationErrors); isValidationErr {
		code = http.StatusUnprocessableEntity

		validationErrMsgs = make([]validationErrMsg, 0)
		for _, err := range validationErrs {
			validationErrMsgs = append(validationErrMsgs, getValidationErrMsg(err))
		}
	}

	if c.Request().Method == http.MethodHead {
		err = c.NoContent(code)
	} else {
		if len(validationErrMsgs) > 0 {
			err = render.Error(c, code, "Validation failed.", validationErrMsgs)
		} else {
			err = render.Error(c, code, err.Error(), nil)
		}
	}
}

type validationErrMsg struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getValidationErrMsg(err validator.FieldError) validationErrMsg {
	errMsg := validationErrMsg{
		Field: err.Field(),
	}

	errMsg.Message = map[string]string{
		"required":         err.Field() + " is required.",
		"required_without": err.Field() + " is required.",
		"email":            "Must be a valid email address.",
		"min":              err.Field() + " must be at least " + err.Param() + " characters long.",
		"max":              err.Field() + " must not exceed " + err.Param() + " characters.",
		"url":              "Must be a valid URL.",
		"oneof":            err.Field() + " must be one of the following: " + err.Param() + ".",
	}[err.Tag()]

	return errMsg
}
