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
	GetUser(ctx context.Context, accountID int) (io.GetUserOutput, error)
	ListUsers(ctx context.Context, input io.ListUsersInput) (io.ListUsersOutput, error)
	DeleteUser(ctx context.Context, accountID int) error
	SuspendUser(ctx context.Context, accountID int) error
	UnsuspendUser(ctx context.Context, accountID int) error

	ListRoadmaps(ctx context.Context, input io.ListRoadmapsInput) (io.ListRoadmapsOutput, error)
	DeleteRoadmap(ctx context.Context, input io.DeleteRoadmapInput) (io.DeleteRoadmapOutput, error)
}

type adminApplication struct {
	repository *repository.Repository
	auth       *auth.Manager
	tracer     trace.Tracer
}

var _ Application = (*adminApplication)(nil)

func New(
	repository *repository.Repository,
	auth *auth.Manager,
	tracer *tracesdk.TracerProvider,
) Application {
	return &adminApplication{
		repository: repository,
		auth:       auth,
		tracer:     tracer.Tracer("admin"),
	}
}
