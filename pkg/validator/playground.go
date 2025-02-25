package validator

import (
	"errors"
	"reflect"
	"strings"

	playground "github.com/go-playground/validator/v10"
)

type PlaygroundValidator struct {
	validator *playground.Validate
}

func NewPlayground() Validator {
	validator := playground.New()

	// https://github.com/go-playground/validator/issues/258#issuecomment-257281334
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &PlaygroundValidator{validator: validator}
}

func (v *PlaygroundValidator) Validate(s any) error {
	return v.validator.Struct(s)
}

func (v *PlaygroundValidator) ParseErrors(err error) []ErrorMessage {
	var errMessages []ErrorMessage
	var errs playground.ValidationErrors
	if isValidationErr := errors.As(err, &errs); isValidationErr {
		errMessages = make([]ErrorMessage, 0)
		for _, err := range errs {
			errMessages = append(errMessages, getValidationErrMsg(err))
		}
	}
	return errMessages
}

func getValidationErrMsg(err playground.FieldError) ErrorMessage {
	errMsg := ErrorMessage{
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
