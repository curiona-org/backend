package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/pkg/server/render"
)

func (h *Handler) GetTopicBySlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return apperrors.NotFound()
	}

	output, err := h.application.GetTopicBySlug(c.Request().Context(), slug)
	if err != nil {
		return err
	}

	return render.OK(c, "Profile details.", output)
}
