package middleware

import (
	"context"
	"strings"

	"github.com/curiona-org/backend/internal/app"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(app app.CurionaApplication) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqCtx := c.Request().Context()
			authorization, ok := c.Request().Header["Authorization"]
			if !ok {
				return cerrors.Unauthorized
			}

			bearer := strings.Split(authorization[0], " ")
			if len(bearer) < 2 {
				return cerrors.Unauthorized
			}

			token := bearer[1]
			payload, err := app.AuthVerify(reqCtx, token)
			if err != nil {
				return cerrors.Unauthorized
			}

			ctx := context.WithValue(reqCtx, auth.ContextKey, payload)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
