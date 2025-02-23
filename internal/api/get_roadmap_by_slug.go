package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/cerrors"
	"github.com/roadmap-thesis/backend/internal/server/render"
)

func (a *API) GetRoadmapBySlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return cerrors.NotFound()
	}

	output, err := a.application.GetRoadmapBySlug(c.Request().Context(), slug)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
