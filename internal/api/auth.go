package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/application/io"
	"github.com/roadmap-thesis/backend/pkg/server/render"
)

func (a *API) Auth(c echo.Context) error {
	var input io.AuthInput

	if err := c.Bind(&input); err != nil {
		return apperrors.InvalidData()
	}

	if err := c.Validate(&input); err != nil {
		return err
	}

	input.ClientIP = c.RealIP()
	input.UserAgent = c.Request().UserAgent()
	output, err := a.application.Auth(c.Request().Context(), input)
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/",
		MaxAge:   output.RefreshTokenExpiresIn,
		Expires:  output.RefreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	if output.Created {
		return render.Created(c, "Successfully registered.", output)
	}

	return render.OK(c, "Successfully logged in.", output)
}
