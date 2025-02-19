package domain

import (
	"context"
	"errors"
	"time"

	"github.com/roadmap-thesis/backend/pkg/slug"
	"github.com/roadmap-thesis/backend/pkg/str"
)

const (
	TopicTable = "topics"
)

var (
	ErrTopicNotFound = errors.New("topic not found")
)

type Topic struct {
	ID                  int
	RoadmapID           int
	ParentID            int
	Title               string
	Slug                string
	Description         string
	Order               int
	Finished            bool
	ExternalSearchQuery string

	Subtopics []*Topic
	Resources []ExternalResource

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TopicRepository interface {
	GetBySlug(ctx context.Context, slug string) (Topic, error)
}

func NewTopic(title, description, externalSearchQuery string) *Topic {
	return &Topic{
		Title:               title,
		Slug:                slug.Make(title + " " + str.Random(5)),
		Description:         description,
		Finished:            false,
		ExternalSearchQuery: externalSearchQuery,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

func (e *Topic) IsZero() bool {
	return e.ID == 0 &&
		e.RoadmapID == 0 &&
		e.ParentID == 0 &&
		e.Title == "" &&
		e.Slug == "" &&
		e.Description == "" &&
		e.Order == 0 &&
		!e.Finished &&
		len(e.Subtopics) == 0 &&
		len(e.Resources) == 0 &&
		e.CreatedAt.IsZero() &&
		e.UpdatedAt.IsZero()
}

func (e *Topic) IsParent() bool {
	return e.ParentID == 0
}

func (e *Topic) IsChild() bool {
	return e.ParentID != 0
}

func (e *Topic) HasSubtopics() bool {
	return len(e.Subtopics) > 0
}

func (e *Topic) HasResources() bool {
	return len(e.Resources) > 0
}

func (e *Topic) GetSubtopic(id int) *Topic {
	for _, subtopic := range e.Subtopics {
		if subtopic.ID == id {
			return subtopic
		}
	}

	return nil
}

func (e *Topic) AddSubtopic(subtopic *Topic) {
	if e.Subtopics == nil {
		e.Subtopics = make([]*Topic, 0)
	}

	subtopic.Order = len(e.Subtopics) + 1

	subtopic.ParentID = e.ID
	e.Subtopics = append(e.Subtopics, subtopic)
}

func (e *Topic) AddResource(resource ...ExternalResource) {
	if e.Resources == nil {
		e.Resources = make([]ExternalResource, 0)
	}

	e.Resources = append(e.Resources, resource...)
}

func (e *Topic) GetYoutubeResources() []ExternalResource {
	var youtubeResources []ExternalResource

	for _, resource := range e.Resources {
		if resource.IsYoutube() {
			youtubeResources = append(youtubeResources, resource)
		}
	}

	return youtubeResources
}

func (e *Topic) GetBookResources() []ExternalResource {
	var bookResources []ExternalResource

	for _, resource := range e.Resources {
		if resource.IsBook() {
			bookResources = append(bookResources, resource)
		}
	}

	return bookResources
}

func (e *Topic) GetArticleResources() []ExternalResource {
	var articleResources []ExternalResource

	for _, resource := range e.Resources {
		if resource.IsArticle() {
			articleResources = append(articleResources, resource)
		}
	}

	return articleResources
}

func (e *Topic) Update(title, description, slug string) {
	e.Title = title
	e.Description = description
	e.Slug = slug
	e.UpdateChangelog()
}

func (e *Topic) MarkAsFinished() {
	e.Finished = true
	e.UpdateChangelog()
}

func (e *Topic) UpdateChangelog() {
	e.UpdatedAt = time.Now()
}
