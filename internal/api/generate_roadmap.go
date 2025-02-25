package api

import (
	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) GenerateRoadmap(c echo.Context) error {
	var input io.GenerateRoadmapInput

	if err := c.Bind(&input); err != nil {
		return cerrors.InvalidData
	}

	if err := c.Validate(&input); err != nil {
		return err
	}

	auth := auth.TokenFromContext(c.Request().Context())
	input.AccountID = auth.AccountID

	output, err := a.application.GenerateRoadmap(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return render.Created(c, "Roadmap generated successfully", output)
}
