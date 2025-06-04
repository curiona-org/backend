package domain

import (
	"errors"
	"time"
)

const (
	RatingTable = "roadmap_ratings"
)

var (
	ErrRatingNotFound = errors.New("rating not found")
)

type Rating struct {
	RoadmapID                      int
	AccountID                      int
	ProgressionTotalTopics         int
	ProgressionTotalFinishedTopics int
	Rating                         int
	Comment                        string

	Account *Account

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRating(roadmapID, accountID, progressionTotalTopics, progressionTotalFinishedTopics, rating int, comment string) *Rating {
	return &Rating{
		RoadmapID:                      roadmapID,
		AccountID:                      accountID,
		ProgressionTotalTopics:         progressionTotalTopics,
		ProgressionTotalFinishedTopics: progressionTotalFinishedTopics,
		Rating:                         rating,
		Comment:                        comment,
		CreatedAt:                      time.Now(),
		UpdatedAt:                      time.Now(),
	}
}

func (r *Rating) IsZero() bool {
	return r.RoadmapID == 0 &&
		r.AccountID == 0 &&
		r.ProgressionTotalTopics == 0 &&
		r.ProgressionTotalFinishedTopics == 0 &&
		r.Rating == 0 &&
		r.Comment == ""
}

func (r *Rating) SetAccount(account *Account) {
	r.Account = account
}
