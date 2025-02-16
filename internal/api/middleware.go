package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func (s *Server) setupMiddlewares() {
	s.instance.Use(middleware.CORS())
	s.instance.Use(middleware.Recover())
	s.instance.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.Debug().
					Str("uri", v.URI).
					Int("status", v.Status).
					Send()
			} else {
				if config.IsProduction() || v.Status >= 500 {
					log.Error().
						Err(v.Error).
						Str("uri", v.URI).
						Int("status", v.Status).
						Send()
				}
			}
			return nil
		},
	}))
	// if config.IsProduction() {
	// s.instance.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	// }
	s.instance.Use(otelecho.Middleware("api"))
}
