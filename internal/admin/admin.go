package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/repository"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Application is the interface that provides the admin application methods.
type Application interface {
	Statistics(ctx context.Context) (io.StatisticsOutput, error)

	IsAdmin(ctx context.Context, accountID int) (bool, error)
	ListUsers(ctx context.Context, input io.ListUsersInput) (io.ListUsersOutput, error)
	DeleteUser(ctx context.Context, input io.DeleteUserInput) (io.DeleteUserOutput, error)

	ListRoadmaps(ctx context.Context, input io.ListRoadmapsInput) (io.ListRoadmapsOutput, error)
	DeleteRoadmap(ctx context.Context, input io.DeleteRoadmapInput) (io.DeleteRoadmapOutput, error)
}

type adminApplication struct {
	repository *repository.Repository
	auth       *auth.Auth
	tracer     trace.Tracer
}

var _ Application = (*adminApplication)(nil)

func New(
	repository *repository.Repository,
	auth *auth.Auth,
	tracer *tracesdk.TracerProvider,
) Application {
	return &adminApplication{
		repository: repository,
		auth:       auth,
		tracer:     tracer.Tracer("admin"),
	}
}
