package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/auth"
	"github.com/roadmap-thesis/backend/internal/server/render"
)

func (a *API) GetProfile(c echo.Context) error {
	auth := auth.FromContext(c.Request().Context())

	output, err := a.application.GetProfile(c.Request().Context(), auth.ID)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
