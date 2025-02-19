package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/server/render"
)

func (a *API) UpdateProfile(c echo.Context) error {
	var input io.UpdateProfileInput

	if err := c.Bind(&input); err != nil {
		return apperrors.InvalidData()
	}

	if err := c.Validate(&input); err != nil {
		return err
	}

	output, err := a.application.UpdateProfile(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
