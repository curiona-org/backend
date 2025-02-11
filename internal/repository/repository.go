package repository

import (
	"github.com/roadmap-thesis/backend/pkg/database"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("repository-layer")
)

type Repository struct {
	Account                *AccountRepository
	Roadmap                *RoadmapRepository
	Topic                  *TopicRepository
	PersonalizationOptions *PersonalizationOptionsRepository
	Session                *SessionRepository
}

func New(db database.Connection) *Repository {
	return &Repository{
		Account:                NewAccountRepository(db),
		Roadmap:                NewRoadmapRepository(db),
		Topic:                  NewTopicRepository(db),
		PersonalizationOptions: NewPersonalizationOptionsRepository(db),
		Session:                NewSessionRepository(db),
	}
}
