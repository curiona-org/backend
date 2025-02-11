package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/io"
	"github.com/roadmap-thesis/backend/pkg/render"
)

func (h *Handler) GenerateRoadmap(c echo.Context) error {
	var input io.GenerateRoadmapInput

	if err := c.Bind(&input); err != nil {
		return apperrors.InvalidData()
	}

	if err := c.Validate(&input); err != nil {
		return err
	}

	output, err := h.application.GenerateRoadmap(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return render.Created(c, "Roadmap generated successfully", output)
}
