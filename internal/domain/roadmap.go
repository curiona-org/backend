package domain

import (
	"errors"
	"time"

	"github.com/curiona-org/backend/pkg/slug"
	"github.com/curiona-org/backend/pkg/str"
)

const (
	RoadmapTable = "roadmaps"

	RoadmapMinimumTopics    = 5
	RoadmapMaximumTopics    = 10
	RoadmapMinimumSubtopics = 3
	RoadmapMaximumSubtopics = 5

	RoadmapTopicProgressionTable = "roadmap_topic_progressions"
)

var (
	ErrRoadmapNotFound = errors.New("roadmap not found")
)

type Roadmap struct {
	ID                  int
	AccountID           int
	Title               string
	Slug                string
	Description         string
	TotalTopics         int
	TotalFinishedTopics int

	Account                *Account
	Topics                 []*Topic
	PersonalizationOptions *PersonalizationOptions

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
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
		e.TotalFinishedTopics == 0 &&
		e.TotalTopics == 0 &&
		e.Account.IsZero() &&
		len(e.Topics) == 0 &&
		e.PersonalizationOptions.IsZero() &&
		e.CreatedAt.IsZero() &&
		e.UpdatedAt.IsZero()
}

func (e *Roadmap) AddTopic(topic *Topic) {
	if e.Topics == nil {
		e.Topics = make([]*Topic, 0)
	}

	topic.Order = len(e.Topics) + 1

	e.Topics = append(e.Topics, topic)
}

func (e *Roadmap) CompletionPercentage() float64 {
	if e.TotalTopics == 0 {
		return 0
	}

	return float64(e.TotalFinishedTopics) / float64(e.TotalTopics) * 100
}

func (e *Roadmap) SetCreator(acc *Account) {
	e.Account = acc
}

func (e *Roadmap) SetTopics(topics []*Topic) {
	e.Topics = topics
}

func (e *Roadmap) SetPersonalizationOptions(opts *PersonalizationOptions) {
	e.PersonalizationOptions = opts
}

func (e *Roadmap) IsDeleted() bool {
	return !e.DeletedAt.IsZero()
}

func (e *Roadmap) Delete() {
	e.DeletedAt = time.Now()
}

func (e *Roadmap) UpdateChangelog() {
	e.UpdatedAt = time.Now()
}

type RoadmapTopicProgression struct {
	ID         int
	AccountID  int
	RoadmapID  int
	TopicID    int
	IsFinished bool
}
