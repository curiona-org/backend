package domain

import (
	"context"
	"errors"
	"time"

	"github.com/roadmap-thesis/backend/internal/domain/object"
)

const (
	ExternalResourceTable = "external_resources"
)

var (
	ErrExternalResourcesNotFound = errors.New("external resources not found")
)

type ExternalResource struct {
	ID      int
	TopicID int
	Title   string
	URL     string
	Type    object.ExternalResourceType

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExternalResourceRepository interface {
	GetByTopicID(ctx context.Context, topicID int) ([]ExternalResource, error)
	BulkSave(ctx context.Context, topicID int, resource []*ExternalResource) error
}

func NewExternalResource(topicID int, title, url string, resourceType object.ExternalResourceType) *ExternalResource {
	return &ExternalResource{
		TopicID:   topicID,
		Title:     title,
		URL:       url,
		Type:      resourceType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (e *ExternalResource) IsYoutube() bool {
	return e.Type == object.ExternalResourceTypeYoutube
}

func (e *ExternalResource) IsBook() bool {
	return e.Type == object.ExternalResourceTypeBook
}

func (e *ExternalResource) IsArticle() bool {
	return e.Type == object.ExternalResourceTypeArticle
}

func (e *ExternalResource) IsZero() bool {
	return e.ID == 0 &&
		e.TopicID == 0 &&
		e.Title == "" &&
		e.URL == "" &&
		e.Type == ""
}
