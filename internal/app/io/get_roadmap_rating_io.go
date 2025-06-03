package io

import "time"

type GetUserRoadmapRatingInput struct {
	AccountID int    `json:"-"`
	Slug      string `json:"slug"`
}

type GetUserRoadmapRatingOutput struct {
	IsRated                        bool      `json:"is_rated"`
	RoadmapID                      int       `json:"roadmap_id"`
	ProgressionTotalTopics         int       `json:"progression_total_topics"`
	ProgressionTotalFinishedTopics int       `json:"progression_total_finished_topics"`
	Rating                         int       `json:"rating"`
	Comment                        string    `json:"comment"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}
