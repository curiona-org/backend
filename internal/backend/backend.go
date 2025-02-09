package backend

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/io"
	"github.com/roadmap-thesis/backend/internal/repository"
	"github.com/roadmap-thesis/backend/pkg/auth"
	"github.com/roadmap-thesis/backend/pkg/auth/oauth"
	"github.com/roadmap-thesis/backend/pkg/llm"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("backend-layer")
)

type Backend interface {
	Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error)
	AuthVerify(ctx context.Context, token string) (*auth.Payload, error)
	GetProfile(ctx context.Context) (io.GetProfileOutput, error)

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

type backend struct {
	repository  *repository.Repository
	llm         llm.Client
	auth        auth.Authorizer
	googleOAuth oauth.Client
}

func New(repository *repository.Repository, llm llm.Client, auth auth.Authorizer, googleOAuth oauth.Client) Backend {
	return &backend{
		repository:  repository,
		llm:         llm,
		auth:        auth,
		googleOAuth: googleOAuth,
	}
}
