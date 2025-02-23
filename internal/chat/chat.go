package chat

import (
	"github.com/curiona-org/backend/internal/repository"
	"github.com/curiona-org/backend/internal/worker"
	"github.com/curiona-org/backend/pkg/auth"
	"github.com/curiona-org/backend/pkg/llm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Application handles chat assistance functionality. It interacts with the LLM stream provider
// to send and receive messages, and with the repository to store and retrieve messages.
type Application interface {
}

type application struct {
	worker     worker.Worker
	repository *repository.Repository
	llm        llm.Client
	auth       *auth.Auth
	tracer     trace.Tracer
}

var _ Application = (*application)(nil)

func New(worker worker.Worker, repository *repository.Repository, llm llm.Client, auth *auth.Auth) Application {
	tracer := otel.Tracer("chat")
	return &application{
		repository: repository,
		llm:        llm,
		auth:       auth,
		worker:     worker,
		tracer:     tracer,
	}
}
