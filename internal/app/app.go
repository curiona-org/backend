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
	HealthCheck(ctx context.Context) bool

	Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error)
	AuthVerify(ctx context.Context, token string) (*auth.Token, error)
	AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error)
	GetProfile(ctx context.Context, accountID int) (io.GetProfileOutput, error)
	UpdateProfile(ctx context.Context, input io.UpdateProfileInput) (io.UpdateProfileOutput, error)

	GetRoadmapBySlug(ctx context.Context, input io.GetRoadmapInput) (io.GetRoadmapOutput, error)
	GenerateRoadmap(ctx context.Context, input io.GenerateRoadmapInput) (io.GenerateRoadmapOutput, error)
	RegenerateRoadmap(ctx context.Context, input io.RegenerateRoadmapInput) (io.RegenerateRoadmapOutput, error)
	ListCommunityRoadmaps(ctx context.Context, input io.ListCommunityRoadmapsInput) (io.ListCommunityRoadmapsOutput, error)
	ListUserRoadmaps(ctx context.Context, input io.ListUserRoadmapsInput) (io.ListUserRoadmapsOutput, error)
	ListUserFinishedRoadmaps(ctx context.Context, input io.ListUserFinishedRoadmapsInput) (io.ListUserFinishedRoadmapsOutput, error)
	ListUserOnProgressRoadmaps(ctx context.Context, input io.ListUserOnProgressRoadmapsInput) (io.ListUserOnProgressRoadmapsOutput, error)
	DeleteUserRoadmap(ctx context.Context, input io.DeleteUserRoadmapInput) error

	GetUserRoadmapRating(ctx context.Context, input io.GetUserRoadmapRatingInput) (io.GetUserRoadmapRatingOutput, error)
	RateRoadmap(ctx context.Context, input io.RateRoadmapInput) error

	// RoadmapChatAssistStream handles chat assistance functionality. It interacts with the LLM stream
	// provider to send and receive messages.
	RoadmapChatAssistStream(ctx context.Context, input io.RoadmapChatAssistStreamInput) (llm.Stream, error)

	GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error)
	MarkTopicAsFinished(ctx context.Context, input io.MarkTopicInput) error
	MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicInput) error

	BookmarkRoadmap(ctx context.Context, input io.BookmarkRoadmapInput) error
	UnbookmarkRoadmap(ctx context.Context, input io.BookmarkRoadmapInput) error
	ListBookmarkedRoadmaps(ctx context.Context, input io.ListBookmarkedRoadmapsInput) (io.ListBookmarkedRoadmapsOutput, error)
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
