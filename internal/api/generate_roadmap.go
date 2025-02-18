package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/pkg/server/render"
)

func (a *API) GenerateRoadmap(c echo.Context) error {
	var input io.GenerateRoadmapInput

	if err := c.Bind(&input); err != nil {
		return apperrors.InvalidData()
	}

	if err := c.Validate(&input); err != nil {
		return err
	}

	output, err := a.application.GenerateRoadmap(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return render.Created(c, "Roadmap generated successfully", output)
}
