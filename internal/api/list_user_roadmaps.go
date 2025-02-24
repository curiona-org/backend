package api

import (
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) ListUserRoadmaps(c echo.Context) error {
	auth := auth.TokenFromContext(c.Request().Context())

	output, err := a.application.ListUserRoadmaps(c.Request().Context(), auth.AccountID)
	if err != nil {
		return err
	}

	return render.OK(c, "List User Roadmaps.", output)
}
