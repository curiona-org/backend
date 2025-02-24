package api

import (
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) GetProfile(c echo.Context) error {
	auth := auth.FromContext(c.Request().Context())

	output, err := a.application.GetProfile(c.Request().Context(), auth.AccountID())
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
