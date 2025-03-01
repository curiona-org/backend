package domain

import (
	"errors"
	"time"
)

const (
	ExternalResourceTable = "external_resources"
)

var (
	ErrExternalResourcesNotFound = errors.New("external resources not found")
)

type ExternalResourceType string

const (
	ExternalResourceTypeYoutube ExternalResourceType = "youtube"
	ExternalResourceTypeBook    ExternalResourceType = "book"
	ExternalResourceTypeArticle ExternalResourceType = "article"
)

type ExternalResource struct {
	ID      int
	TopicID int
	Title   string
	URL     string
	Type    ExternalResourceType

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewExternalResource(topicID int, title, url string, resourceType ExternalResourceType) *ExternalResource {
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
	return e.Type == ExternalResourceTypeYoutube
}

func (e *ExternalResource) IsBook() bool {
	return e.Type == ExternalResourceTypeBook
}

func (e *ExternalResource) IsArticle() bool {
	return e.Type == ExternalResourceTypeArticle
}

func (e *ExternalResource) IsZero() bool {
	return e.ID == 0 &&
		e.TopicID == 0 &&
		e.Title == "" &&
		e.URL == "" &&
		e.Type == ""
}
