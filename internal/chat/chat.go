package chat

import (
	"github.com/roadmap-thesis/backend/internal/auth"
	"github.com/roadmap-thesis/backend/internal/llm"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/internal/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ChatApp handles assistance chat functionality. It interacts with the LLM stream provider
// to send and receive messages, and with the repository to store and retrieve messages.
type ChatApp interface {
}

type application struct {
	repository repository.Repository
	llm        llm.Client
	auth       *auth.Auth
	worker     worker.Worker
	tracer     trace.Tracer
}

var _ ChatApp = (*application)(nil)

func New(repository repository.Repository, auth *auth.Auth) ChatApp {
	tracer := otel.Tracer("chat")
	return &application{
		repository: repository,
		auth:       auth,
		tracer:     tracer,
	}
}
