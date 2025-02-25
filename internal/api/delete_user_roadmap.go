package api

import (
	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) DeleteUserRoadmap(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return cerrors.ErrNotFound
	}

	ctx := c.Request().Context()
	auth := auth.TokenFromContext(ctx)
	err := a.application.DeleteUserRoadmap(ctx, io.DeleteUserRoadmapInput{
		AccountID: auth.AccountID,
		Slug:      slug,
	})
	if err != nil {
		return err
	}

	return render.OK(c, "Roadmap deleted.", nil)
}
