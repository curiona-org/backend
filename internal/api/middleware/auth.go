package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/app"
	"github.com/roadmap-thesis/backend/internal/auth"
	"github.com/roadmap-thesis/backend/internal/cerrors"
)

func AuthMiddleware(app app.Application) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqCtx := c.Request().Context()
			authorization, ok := c.Request().Header["Authorization"]
			if !ok {
				return cerrors.Unauthorized()
			}

			bearer := strings.Split(authorization[0], " ")
			if len(bearer) < 2 {
				return cerrors.Unauthorized()
			}

			token := bearer[1]
			payload, err := app.AuthVerify(reqCtx, token)
			if err != nil {
				return cerrors.Unauthorized()
			}

			ctx := context.WithValue(reqCtx, auth.AuthCtxKey, payload)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
