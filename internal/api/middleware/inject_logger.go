package middleware

import (
	"github.com/curiona-org/backend/internal/logger"
	"github.com/labstack/echo/v4"
)

func InjectLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := logger.Get()
		ctx := log.WithContext(c.Request().Context())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
