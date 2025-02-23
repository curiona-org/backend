package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/pkg/auth"
	"github.com/roadmap-thesis/backend/pkg/cerrors"
	"github.com/roadmap-thesis/backend/pkg/server/render"
)

func (a *API) MarkTopicAsIncomplete(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return cerrors.NotFound()
	}

	auth := auth.FromContext(c.Request().Context())
	err := a.application.MarkTopicAsIncomplete(c.Request().Context(), io.MarkTopicInput{
		Slug:      slug,
		AccountID: auth.ID,
	})
	if err != nil {
		return err
	}

	return render.OK(c, "Topic marked as incomplete.", nil)
}
