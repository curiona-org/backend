package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/pkg/render"
)

func (h *Handler) GetRoadmapBySlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return apperrors.NotFound()
	}

	output, err := h.application.GetRoadmapBySlug(c.Request().Context(), slug)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
