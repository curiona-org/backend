package render

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func OK(c echo.Context, message string, data any) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c echo.Context, message string, data any) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c echo.Context, code int, message string, err any) error {
	return c.JSON(code, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}
