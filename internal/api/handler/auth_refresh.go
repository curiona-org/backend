package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/io"
	"github.com/roadmap-thesis/backend/pkg/apperrors"
	"github.com/roadmap-thesis/backend/pkg/render"
)

func (h *Handler) AuthRefresh(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return apperrors.Unauthorized()
	}

	output, err := h.application.AuthRefresh(c.Request().Context(), io.AuthRefreshInput{
		Token: refreshToken.Value,
	})
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/",
		Expires:  output.RefreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	return render.OK(c, "Successfully refreshed token.", output)
}
