package api

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/application"
	"github.com/roadmap-thesis/backend/pkg/auth"
	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/roadmap-thesis/backend/pkg/server"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"golang.org/x/time/rate"
)

type API struct {
	instance    *server.Server
	application application.Application
}

func New(port string, application application.Application) *API {
	instance := server.New(port)

	api := &API{
		application: application,
		instance:    instance,
	}

	api.setupMiddlewares()
	api.setupRoutes()
	api.instance.HTTPErrorHandler = api.ErrorHandler

	return api
}

func (a *API) Start(ctx context.Context) {
	exit := a.instance.Listen()

	signal := <-exit
	a.instance.Shutdown(ctx, signal)
}

func (a *API) setupRoutes() {
	a.instance.GET("/health", a.HealthCheck)

	a.instance.POST("/auth", a.Auth)
	a.instance.POST("/auth/refresh", a.AuthRefresh)

	a.instance.GET("/profile", a.GetProfile, a.authMiddleware)
	a.instance.PATCH("/profile", a.UpdateProfile, a.authMiddleware)

	a.instance.GET("/roadmaps", a.ListUserRoadmaps, a.authMiddleware)
	a.instance.GET("/roadmaps/:slug", a.GetRoadmapBySlug, a.authMiddleware)
	a.instance.POST("/roadmaps", a.GenerateRoadmap, a.authMiddleware)
	a.instance.GET("/roadmaps/topic/:slug", a.GetTopicBySlug, a.authMiddleware)
}

func (a *API) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqCtx := c.Request().Context()
		authorization, ok := c.Request().Header["Authorization"]
		if !ok {
			return apperrors.Unauthorized()
		}

		bearer := strings.Split(authorization[0], " ")
		if len(bearer) < 2 {
			return apperrors.Unauthorized()
		}

		token := bearer[1]
		payload, err := a.application.AuthVerify(reqCtx, token)
		if err != nil {
			return apperrors.Unauthorized()
		}

		ctx := context.WithValue(reqCtx, auth.AuthCtxKey, payload)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func (a *API) setupMiddlewares() {
	a.instance.Use(middleware.CORS())
	a.instance.Use(middleware.Recover())
	a.instance.Use(middleware.RequestID())
	a.instance.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error { //nolint:revive
			if v.Error == nil {
				log.Debug().
					Str("uri", v.URI).
					Int("status", v.Status).
					Send()
			} else if config.IsProduction() || v.Status >= 500 {
				log.Error().
					Err(v.Error).
					Str("uri", v.URI).
					Int("status", v.Status).
					Send()
			}

			return nil
		},
	}))
	if config.IsProduction() {
		a.instance.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	}
	a.instance.Use(otelecho.Middleware("api"))
}
