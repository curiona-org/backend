package api

import (
	"context"

	"github.com/curiona-org/backend/internal/admin"
	"github.com/curiona-org/backend/internal/api/middleware"
	"github.com/curiona-org/backend/internal/app"
	"github.com/curiona-org/backend/internal/chat"
	"github.com/curiona-org/backend/internal/config"
	"github.com/curiona-org/backend/pkg/server"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

type API struct {
	instance    *server.Server
	application app.CurionaApplication
	adminApp    admin.Application
	chatApp     chat.Application
}

func New(port string, curionaApp app.CurionaApplication, adminApp admin.Application, chatApp chat.Application) *API {
	instance := server.New(port)

	api := &API{
		instance:    instance,
		application: curionaApp,
		adminApp:    adminApp,
		chatApp:     chatApp,
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

	authMiddleware := middleware.AuthMiddleware(a.application)
	a.instance.GET("/profile", a.GetProfile, authMiddleware)
	a.instance.PATCH("/profile", a.UpdateProfile, authMiddleware)

	a.instance.GET("/roadmaps", a.ListUserRoadmaps, authMiddleware)
	a.instance.GET("/roadmaps/:slug", a.GetRoadmapBySlug, authMiddleware)
	a.instance.POST("/roadmaps", a.GenerateRoadmap, authMiddleware)
	a.instance.DELETE("/roadmaps/:slug", a.DeleteUserRoadmap, authMiddleware)
	a.instance.GET("/roadmaps/topic/:slug", a.GetTopicBySlug, authMiddleware)
	a.instance.PATCH("/roadmaps/topic/:slug/finish", a.MarkTopicAsFinished, authMiddleware)
	a.instance.PATCH("/roadmaps/topic/:slug/incomplete", a.MarkTopicAsIncomplete, authMiddleware)
}

func (a *API) setupMiddlewares() {
	a.instance.Use(echoMiddleware.CORS())
	a.instance.Use(echoMiddleware.Recover())
	a.instance.Use(echoMiddleware.RequestID())
	a.instance.Use(echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error { //nolint:revive
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
	// if config.IsProduction() {
	// 	a.instance.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	// }
	a.instance.Use(otelecho.Middleware("api"))
}
