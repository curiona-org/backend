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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Application interface {
	Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error)
	AuthVerify(ctx context.Context, token string) (*auth.Payload, error)
	AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error)
	GetProfile(ctx context.Context) (io.GetProfileOutput, error)
	UpdateProfile(ctx context.Context, input io.UpdateProfileInput) (io.UpdateProfileOutput, error)

	GetRoadmapBySlug(ctx context.Context, slug string) (io.GetRoadmapOutput, error)
	GenerateRoadmap(ctx context.Context, input io.GenerateRoadmapInput) (io.GenerateRoadmapOutput, error)
	ListUserRoadmaps(ctx context.Context) (io.ListUserRoadmapsOutput, error)

	GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error)
	// DeleteUserRoadmap(ctx context.Context, input io.DeleteUserRoadmapInput) (io.DeleteUserRoadmapOutput, error)
	// RegenerateRoadmap(ctx context.Context, input io.RegenerateRoadmapInput) (io.RegenerateRoadmapOutput, error)
	// GetTopicResources(ctx context.Context, input io.GetTopicResourcesInput) (io.GetTopicResourcesOutput, error)
	// MarkTopicAsFinish(ctx context.Context, input io.MarkTopicAsFinishInput) (io.TopicFinishOutput, error)
	// MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicAsIncompleteInput) (io.TopicFinishOutput, error)
}

type application struct {
	repository  *repository.Repository
	llm         llm.Client
	auth        *auth.Auth
	googleOAuth oauth.Client
	googleBooks book.Client
	youtube     youtube.Client
	tracer      trace.Tracer
}

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
