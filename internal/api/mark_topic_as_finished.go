package api

import (
	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/pkg/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) MarkTopicAsFinished(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return cerrors.NotFound()
	}

	auth := auth.TokenFromContext(c.Request().Context())

	err := a.application.MarkTopicAsFinished(c.Request().Context(), io.MarkTopicInput{
		Slug:      slug,
		AccountID: auth.AccountID,
	})
	if err != nil {
		return err
	}

	return render.OK(c, "Topic marked as finished.", nil)
}
