package app

import (
	"context"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/auth/oauth"
	"github.com/curiona-org/backend/internal/repository"
	"github.com/curiona-org/backend/internal/worker"
	"github.com/curiona-org/backend/pkg/googleapi/book"
	"github.com/curiona-org/backend/pkg/googleapi/youtube"
	"github.com/curiona-org/backend/pkg/llm"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// CurionaApplication is the main application. Handling authentication, roadmap/topic management, and user profile management.
type CurionaApplication interface {
	Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error)
	AuthVerify(ctx context.Context, token string) (*auth.Token, error)
	AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error)
	GetProfile(ctx context.Context, accountID int) (io.GetProfileOutput, error)
	UpdateProfile(ctx context.Context, input io.UpdateProfileInput) (io.UpdateProfileOutput, error)

	GetRoadmapBySlug(ctx context.Context, slug string) (io.GetRoadmapOutput, error)
	GenerateRoadmap(ctx context.Context, input io.GenerateRoadmapInput) (io.GenerateRoadmapOutput, error)
	// RegenerateRoadmap(ctx context.Context, input io.RegenerateRoadmapInput) (io.RegenerateRoadmapOutput, error)
	ListCommunityRoadmaps(ctx context.Context, input io.ListCommunityRoadmapsInput) (io.ListCommunityRoadmapsOutput, error)
	ListUserRoadmaps(ctx context.Context, accountID int) (io.ListUserRoadmapsOutput, error)
	DeleteUserRoadmap(ctx context.Context, input io.DeleteUserRoadmapInput) error

	// StreamRoadmapLLM handles chat assistance functionality. It interacts with the LLM stream
	// provider to send and receive messages.
	StreamRoadmapLLM(ctx context.Context, input io.StreamRoadmapLLMInput) (llm.Stream, error)

	GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error)
	MarkTopicAsFinished(ctx context.Context, input io.MarkTopicInput) error
	MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicInput) error
}

type application struct {
	worker      worker.Worker
	repository  *repository.Repository
	llm         llm.Client
	auth        *auth.Manager
	googleOAuth oauth.Client
	googleBooks book.Client
	youtube     youtube.Client
	tracer      trace.Tracer
}

var _ CurionaApplication = (*application)(nil)

func New(
	worker worker.Worker,
	repository *repository.Repository,
	llm llm.Client,
	auth *auth.Manager,
	googleOAuth oauth.Client,
	googleBooks book.Client,
	youtube youtube.Client,
	tracer *tracesdk.TracerProvider,
) CurionaApplication {
	return &application{
		worker:      worker,
		repository:  repository,
		llm:         llm,
		auth:        auth,
		googleOAuth: googleOAuth,
		googleBooks: googleBooks,
		youtube:     youtube,
		tracer:      tracer.Tracer("app"),
	}
}
