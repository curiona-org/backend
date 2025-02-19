package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/server/render"
)

func (a *API) ListUserRoadmaps(c echo.Context) error {
	output, err := a.application.ListUserRoadmaps(c.Request().Context())
	if err != nil {
		return err
	}

	return render.OK(c, "List User Roadmaps.", output)
}
