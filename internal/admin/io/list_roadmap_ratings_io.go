package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/filter"
)

type ListRoadmapRatingsInput = filter.Params
type ListRoadmapRatingsOutput = filter.FilteredList[ListRoadmapRatingsOutputItem]

type ListRoadmapRatingsOutputItem struct {
	IsRated                        bool   `json:"is_rated"`
	AccountID                      int    `json:"account_id"`
	RoadmapID                      int    `json:"roadmap_id"`
	ProgressionTotalTopics         int    `json:"progression_total_topics"`
	ProgressionTotalFinishedTopics int    `json:"progression_total_finished_topics"`
	Rating                         int    `json:"rating"`
	Comment                        string `json:"comment"`

	User ListRoadmapRatingsOutputItemUser `json:"user"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListRoadmapRatingsOutputItemUser struct {
	ID          int         `json:"id"`
	Method      auth.Method `json:"method"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Avatar      string      `json:"avatar"`
	IsSuspended bool        `json:"is_suspended"`
	JoinedAt    time.Time   `json:"joined_at"`
}
