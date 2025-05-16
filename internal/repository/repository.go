package repository

import (
	"context"

	"github.com/curiona-org/backend/pkg/cache"
	"github.com/curiona-org/backend/pkg/database"
)

type Repository struct {
	db                     database.Connection
	Account                *AccountRepository
	Profile                *ProfileRepository
	Roadmap                *RoadmapRepository
	Topic                  *TopicRepository
	ExternalResource       *ExternalResourceRepository
	PersonalizationOptions *PersonalizationOptionsRepository
	Bookmark               *BookmarkRepository
	Session                *SessionRepository
}

func NewPostgresRepository(db database.Connection, cache *cache.Connection) *Repository {
	return &Repository{
		db:                     db,
		Account:                NewPostgresAccountRepository(db),
		Profile:                NewPostgresProfileRepository(db),
		Roadmap:                NewPostgresRoadmapRepository(db, cache),
		Topic:                  NewPostgresTopicRepository(db, cache),
		ExternalResource:       NewPostgresExternalResourceRepository(db, cache),
		PersonalizationOptions: NewPostgresPersonalizationOptionsRepository(db),
		Bookmark:               NewPostgresBookmarkRepository(db, cache),
		Session:                NewPostgresSessionRepository(db),
	}
}

func (r *Repository) Ping(ctx context.Context) bool {
	return r.db.Ping(ctx)
}
