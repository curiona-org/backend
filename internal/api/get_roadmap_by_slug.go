package api

import (
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) GetRoadmapBySlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return cerrors.ErrNotFound
	}

	output, err := a.application.GetRoadmapBySlug(c.Request().Context(), slug)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
