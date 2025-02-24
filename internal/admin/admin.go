package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Application is the interface that provides the admin application methods.
type Application interface {
	Statistics(ctx context.Context) (io.StatisticsOutput, error)

	ListUsers(ctx context.Context) (io.ListUsersOutput, error)
	EditUser(ctx context.Context, input io.EditUserInput) (io.EditUserOutput, error)
	DeleteUser(ctx context.Context, input io.DeleteUserInput) (io.DeleteUserOutput, error)

	ListRoadmaps(ctx context.Context) (io.ListRoadmapsOutput, error)
	DeleteRoadmap(ctx context.Context, input io.DeleteRoadmapInput) (io.DeleteRoadmapOutput, error)
}

type application struct {
	repository *repository.Repository
	auth       *auth.Auth
	tracer     trace.Tracer
}

var _ Application = (*application)(nil)

func New(repository *repository.Repository, auth *auth.Auth) Application {
	tracer := otel.Tracer("admin")
	return &application{
		repository: repository,
		auth:       auth,
		tracer:     tracer,
	}
}
