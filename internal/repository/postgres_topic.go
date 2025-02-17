package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/database"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type topicRepository struct {
	db     database.Connection
	tracer trace.Tracer
}

var _ domain.TopicRepository = (*topicRepository)(nil)

func NewPostgresTopicRepository(db database.Connection) domain.TopicRepository {
	tracer := otel.Tracer("db:postgres:topics")
	return &topicRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *topicRepository) GetBySlug(ctx context.Context, slug string) (domain.Topic, error) {
	query, args := psql.Select(
		sm.Columns(
			psql.Quote(domain.TopicTable, "id"),
			psql.Quote(domain.TopicTable, "roadmap_id"),
			psql.Quote(domain.TopicTable, "parent_id"),
			psql.Quote(domain.TopicTable, "title"),
			psql.Quote(domain.TopicTable, "slug"),
			psql.Quote(domain.TopicTable, "description"),
			psql.Quote(domain.TopicTable, "order"),
			psql.Quote(domain.TopicTable, "finished"),
			psql.Quote(domain.TopicTable, "created_at"),
			psql.Quote(domain.TopicTable, "updated_at"),
		),
		sm.From(domain.TopicTable),
		sm.Where(psql.Quote(domain.TopicTable, "slug").EQ(psql.Arg(slug))),
	).MustBuild(ctx)

	topics, err := r.fetch(ctx, query, args...)
	if err != nil {
		return domain.Topic{}, err
	}

	if len(topics) == 0 {
		return domain.Topic{}, domain.ErrTopicNotFound
	}

	return topics[0], nil
}

func (r *topicRepository) fetch(ctx context.Context, query string, args ...any) ([]domain.Topic, error) {
	ctx, span := spanWithSelectQuery(ctx, r.tracer, "(*topicRepository.fetch)", query)
	defer span.End()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch topics")
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var topics []domain.Topic
	for rows.Next() {
		var topic domain.Topic
		var topicParentID pgtype.Int4
		err = rows.Scan(
			&topic.ID,
			&topic.RoadmapID,
			&topicParentID,
			&topic.Title,
			&topic.Slug,
			&topic.Description,
			&topic.Order,
			&topic.Finished,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if topicParentID.Valid {
			topic.ParentID = int(topicParentID.Int32)
		} else {
			topic.ParentID = 0
		}

		topics = append(topics, topic)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}
