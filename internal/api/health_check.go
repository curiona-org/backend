package api

import (
	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/pkg/server/render"
)

func (a *api) HealthCheck(c echo.Context) error {
	return render.OK(c, "OK", nil)
}
