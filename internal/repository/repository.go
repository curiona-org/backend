package repository

import (
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
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

func NewPostgresRepository(db database.Connection, cacheConn cache.Connection) *Repository {
	return &Repository{
		Account:                NewPostgresAccountRepository(db),
		Profile:                NewPostgresProfileRepository(db),
		Roadmap:                NewPostgresRoadmapRepository(db, cacheConn),
		Topic:                  NewPostgresTopicRepository(db, cacheConn),
		ExternalResource:       NewPostgresExternalResourceRepository(db, cacheConn),
		PersonalizationOptions: NewPostgresPersonalizationOptionsRepository(db),
		Session:                NewPostgresSessionRepository(db),
	}
}
