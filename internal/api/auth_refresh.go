package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) AuthRefresh(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return cerrors.Unauthorized()
	}

	output, err := a.application.AuthRefresh(c.Request().Context(), io.AuthRefreshInput{
		Token:     refreshToken.Value,
		UserAgent: c.Request().UserAgent(),
		ClientIP:  c.RealIP(),
	})
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

	return render.OK(c, "Successfully refreshed token.", output)
}
