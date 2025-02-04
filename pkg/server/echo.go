package server

import (
	"reflect"
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewEchoInstance() *echo.Echo {
	instance := echo.New()
	instance.HideBanner = true
	instance.HidePort = true
	validator := validator.New()

	// https://github.com/go-playground/validator/issues/258#issuecomment-257281334
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	instance.Validator = &CustomValidator{validator: validator}
	return instance
}
