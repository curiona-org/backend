package api

import (
	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/pkg/auth"
	"github.com/curiona-org/backend/pkg/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) DeleteUserRoadmap(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return cerrors.NotFound()
	}

	auth := auth.FromContext(c.Request().Context())
	err := a.application.DeleteUserRoadmap(c.Request().Context(), io.DeleteUserRoadmapInput{
		AccountID: auth.ID,
		Slug:      slug,
	})
	if err != nil {
		return err
	}

	return render.OK(c, "Roadmap deleted.", nil)
}
