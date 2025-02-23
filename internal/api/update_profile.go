package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/auth"
	"github.com/roadmap-thesis/backend/internal/cerrors"
	"github.com/roadmap-thesis/backend/internal/server/render"
)

func (a *API) UpdateProfile(c echo.Context) error {
	var input io.UpdateProfileInput

	if err := c.Bind(&input); err != nil {
		return cerrors.InvalidData()
	}

	if err := c.Validate(&input); err != nil {
		return err
	}

	auth := auth.FromContext(c.Request().Context())
	input.AccountID = auth.ID

	output, err := a.application.UpdateProfile(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
