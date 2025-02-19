package domain

import (
	"errors"
	"time"

	"github.com/roadmap-thesis/backend/internal/slug"
	"github.com/roadmap-thesis/backend/internal/str"
)

const (
	RoadmapTable = "roadmaps"

	RoadmapMinimumTopics    = 5
	RoadmapMaximumTopics    = 10
	RoadmapMinimumSubtopics = 3
	RoadmapMaximumSubtopics = 5
)

var (
	ErrRoadmapNotFound = errors.New("roadmap not found")
)

type Roadmap struct {
	ID          int
	AccountID   int
	Title       string
	Slug        string
	Description string

	Account                Account
	Topics                 []*Topic
	PersonalizationOptions *PersonalizationOptions

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRoadmap(accountID int, title, description string) *Roadmap {
	return &Roadmap{
		AccountID:   accountID,
		Title:       title,
		Slug:        slug.Make(title + " " + str.Random(5)),
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (e *Roadmap) IsZero() bool {
	return e.ID == 0 &&
		e.AccountID == 0 &&
		e.Title == "" &&
		e.Slug == "" &&
		e.Description == "" &&
		e.Account.IsZero() &&
		len(e.Topics) == 0 &&
		e.PersonalizationOptions.IsZero() &&
		e.CreatedAt.IsZero() &&
		e.UpdatedAt.IsZero()
}

func (e *Roadmap) TotalTopics() int {
	total := len(e.Topics)
	for _, topic := range e.Topics {
		subtopicsTotal := len(topic.Subtopics)
		if subtopicsTotal > 0 {
			total += subtopicsTotal
		}
	}
	return total
}

func (e *Roadmap) AddTopic(topic *Topic) {
	if e.Topics == nil {
		e.Topics = make([]*Topic, 0)
	}

	topic.Order = len(e.Topics) + 1

	e.Topics = append(e.Topics, topic)
}

func (e *Roadmap) CompletionPercentage() float64 {
	return e.calculateCompletionPercentage(e.Topics, e.TotalTopics())
}

func (e *Roadmap) calculateCompletionPercentage(topics []*Topic, totalTopics int) float64 {
	if totalTopics == 0 {
		return 0
	}

	totalTopicsFinished := float64(0)
	for _, topic := range topics {
		if len(topic.Subtopics) > 0 {
			totalTopicsFinished += e.calculateCompletionPercentage(topic.Subtopics, totalTopics)
		}

		if topic.Finished {
			totalTopicsFinished++
		}
	}

	return totalTopicsFinished / float64(totalTopics)
}

func (e *Roadmap) SetCreator(acc Account) {
	e.Account = acc
}

func (e *Roadmap) SetTopics(topics []*Topic) {
	e.Topics = topics
}

func (e *Roadmap) SetPersonalizationOptions(opts *PersonalizationOptions) {
	e.PersonalizationOptions = opts
}

func (e *Roadmap) UpdateChangelog() {
	e.UpdatedAt = time.Now()
}
