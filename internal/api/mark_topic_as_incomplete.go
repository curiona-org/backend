package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/server/render"
)

func (a *API) MarkTopicAsIncomplete(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return apperrors.NotFound()
	}

	err := a.application.MarkTopicAsIncomplete(c.Request().Context(), slug)
	if err != nil {
		return err
	}

	return render.OK(c, "Topic marked as incomplete.", nil)
}
