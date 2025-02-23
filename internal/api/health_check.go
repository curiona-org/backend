package api

import (
	"github.com/curiona-org/backend/pkg/server/render"
	"github.com/labstack/echo/v4"
)

func (a *API) HealthCheck(c echo.Context) error {
	return render.OK(c, "OK", nil)
}
