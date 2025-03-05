package repository

import (
	"github.com/curiona-org/backend/pkg/database"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type PersonalizationOptionsRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

// Currently unused.
func NewPostgresPersonalizationOptionsRepository(db database.Connection) *PersonalizationOptionsRepository {
	tracer := otel.Tracer("db:postgres:personalization_options")
	return &PersonalizationOptionsRepository{
		db:     db,
		tracer: tracer,
	}
}
