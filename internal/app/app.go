package app

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/auth"
	"github.com/roadmap-thesis/backend/internal/auth/oauth"
	"github.com/roadmap-thesis/backend/internal/googleapi/book"
	"github.com/roadmap-thesis/backend/internal/googleapi/youtube"
	"github.com/roadmap-thesis/backend/internal/llm"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/internal/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Application interface {
	Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error)
	AuthVerify(ctx context.Context, token string) (*auth.Payload, error)
	AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error)
	GetProfile(ctx context.Context, accountID int) (io.GetProfileOutput, error)
	UpdateProfile(ctx context.Context, input io.UpdateProfileInput) (io.UpdateProfileOutput, error)

	GetRoadmapBySlug(ctx context.Context, slug string) (io.GetRoadmapOutput, error)
	GenerateRoadmap(ctx context.Context, input io.GenerateRoadmapInput) (io.GenerateRoadmapOutput, error)
	ListUserRoadmaps(ctx context.Context, accountID int) (io.ListUserRoadmapsOutput, error)
	// DeleteUserRoadmap(ctx context.Context, input io.DeleteUserRoadmapInput) (io.DeleteUserRoadmapOutput, error)
	// RegenerateRoadmap(ctx context.Context, input io.RegenerateRoadmapInput) (io.RegenerateRoadmapOutput, error)

	GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error)
	MarkTopicAsFinished(ctx context.Context, input io.MarkTopicInput) error
	MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicInput) error
}

type application struct {
	repository  *repository.Repository
	llm         llm.Client
	auth        *auth.Auth
	googleOAuth oauth.Client
	googleBooks book.Client
	youtube     youtube.Client
	worker      worker.Worker
	tracer      trace.Tracer
}

var _ Application = (*application)(nil)

func New(
	repository *repository.Repository,
	llm llm.Client,
	auth *auth.Auth,
	googleOAuth oauth.Client,
	googleBooks book.Client,
	youtube youtube.Client,
) Application {
	tracer := otel.Tracer("app")
	return &application{
		repository:  repository,
		llm:         llm,
		auth:        auth,
		googleOAuth: googleOAuth,
		googleBooks: googleBooks,
		youtube:     youtube,
		tracer:      tracer,
	}
}
