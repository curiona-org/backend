package repository

import (
	"github.com/roadmap-thesis/backend/internal/cache"
	"github.com/roadmap-thesis/backend/internal/database"
)

type Repository struct {
	Account                *AccountRepository
	Profile                *ProfileRepository
	Roadmap                *RoadmapRepository
	Topic                  *TopicRepository
	ExternalResource       *ExternalResourceRepository
	PersonalizationOptions *PersonalizationOptionsRepository
	Session                *SessionRepository
}

func NewPostgresRepository(db database.Connection, cache *cache.Connection) *Repository {
	return &Repository{
		Account:                NewPostgresAccountRepository(db),
		Profile:                NewPostgresProfileRepository(db),
		Roadmap:                NewPostgresRoadmapRepository(db, cache),
		Topic:                  NewPostgresTopicRepository(db, cache),
		ExternalResource:       NewPostgresExternalResourceRepository(db, cache),
		PersonalizationOptions: NewPostgresPersonalizationOptionsRepository(db),
		Session:                NewPostgresSessionRepository(db),
	}
}
