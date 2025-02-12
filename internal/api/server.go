package api

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/api/handler"
	"github.com/roadmap-thesis/backend/internal/application"
	"github.com/roadmap-thesis/backend/pkg/server"
)

type Server struct {
	instance    *server.Server
	application application.Application
	handler     *handler.Handler
}

func New(port string, application application.Application) *Server {
	instance := server.New(port)

	handler := handler.New(application)
	api := &Server{
		application: application,
		handler:     handler,
		instance:    instance,
	}

	api.setupMiddlewares()
	api.setupRoutes()
	api.instance.HTTPErrorHandler = handler.ErrorHandler

	return api
}

func (s *Server) Start(ctx context.Context) {
	exit := s.instance.Listen()

	signal := <-exit
	s.instance.Shutdown(ctx, signal)
}
