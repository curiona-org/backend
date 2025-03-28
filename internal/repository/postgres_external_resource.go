package repository

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ExternalResourceRepository struct {
	db     database.Connection
	cache  *cache.Connection
	tracer trace.Tracer
}

func NewPostgresExternalResourceRepository(db database.Connection, cache *cache.Connection) *ExternalResourceRepository {
	tracer := otel.Tracer("db:postgres:external_resources")
	return &ExternalResourceRepository{
		db:     db,
		cache:  cache,
		tracer: tracer,
	}
}

func (r *ExternalResourceRepository) BulkSave(ctx context.Context, topicID int, resource []*domain.ExternalResource) error {
	ctx, span := r.tracer.Start(ctx, "(*ExternalResourceRepository.BulkSave)")
	defer span.End()

	err := r.db.InTx(ctx, func(tx pgx.Tx) error {
		var resources [][]any

		for _, res := range resource {
			resources = append(resources, []any{
				topicID, res.Title, res.Author, res.URL, res.CoverURL, res.Length, res.Type, res.CreatedAt, res.UpdatedAt,
			})
		}

		// if there are no resources to save, return
		if len(resources) == 0 {
			return nil
		}

		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{domain.ExternalResourceTable},
			[]string{"topic_id", "title", "author", "url", "cover_url", "length", "type", "created_at", "updated_at"},
			pgx.CopyFromRows(resources),
		)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
