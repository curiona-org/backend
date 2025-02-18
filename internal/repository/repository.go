package repository

import (
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/cache"
	"github.com/roadmap-thesis/backend/pkg/database"
)

type Repository interface {
	Account() domain.AccountRepository
	Profile() domain.ProfileRepository
	Roadmap() domain.RoadmapRepository
	Topic() domain.TopicRepository
	ExternalResource() domain.ExternalResourceRepository
	PersonalizationOptions() domain.PersonalizationOptionsRepository
	Session() domain.SessionRepository
}

type repository struct {
	account                domain.AccountRepository
	profile                domain.ProfileRepository
	roadmap                domain.RoadmapRepository
	topic                  domain.TopicRepository
	externalResource       domain.ExternalResourceRepository
	personalizationOptions domain.PersonalizationOptionsRepository
	session                domain.SessionRepository
}

var _ Repository = (*repository)(nil)

func NewPostgresRepository(db database.Connection, cacheConn cache.Connection) Repository {
	return &repository{
		account:                NewPostgresAccountRepository(db),
		profile:                NewPostgresProfileRepository(db),
		roadmap:                NewPostgresRoadmapRepository(db, cacheConn),
		topic:                  NewPostgresTopicRepository(db, cacheConn),
		externalResource:       NewPostgresExternalResourceRepository(db, cacheConn),
		personalizationOptions: NewPostgresPersonalizationOptionsRepository(db),
		session:                NewPostgresSessionRepository(db),
	}
}

func (r *repository) Account() domain.AccountRepository {
	return r.account
}

func (r *repository) Profile() domain.ProfileRepository {
	return r.profile
}

func (r *repository) Roadmap() domain.RoadmapRepository {
	return r.roadmap
}

func (r *repository) Topic() domain.TopicRepository {
	return r.topic
}

func (r *repository) ExternalResource() domain.ExternalResourceRepository {
	return r.externalResource
}

func (r *repository) PersonalizationOptions() domain.PersonalizationOptionsRepository {
	return r.personalizationOptions
}

func (r *repository) Session() domain.SessionRepository {
	return r.session
}
