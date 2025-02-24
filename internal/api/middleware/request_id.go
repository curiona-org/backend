package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
)

var (
	requestLoggerHeader = echo.HeaderXRequestID
)

func RequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		requestID := req.Header.Get(requestLoggerHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}
		res.Header().Set(requestLoggerHeader, requestID)

		ctx := context.WithValue(c.Request().Context(), "request_id", requestID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
