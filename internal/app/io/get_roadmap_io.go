package io

import (
	"time"

	"github.com/curiona-org/backend/pkg/interval"
)

type GetRoadmapInput struct {
	AccountID int
	Slug      string
}

type GetRoadmapOutput struct {
	ID                  int                                    `json:"id"`
	Title               string                                 `json:"title"`
	Slug                string                                 `json:"slug"`
	Description         string                                 `json:"description"`
	TotalTopics         int                                    `json:"total_topics"`
	CreatedAt           time.Time                              `json:"created_at"`
	UpdatedAt           time.Time                              `json:"updated_at"`
	IsBookmarked        bool                                   `json:"is_bookmarked"`
	Progression         GetRoadmapOutputProgression            `json:"progression"`
	Rating              GetRoadmapOutputRating                 `json:"rating"`
	Creator             GetRoadmapOutputCreator                `json:"creator"`
	PersonalizationOpts GetRoadmapOutputPersonalizationOptions `json:"personalization_options"`
	Topics              []GetRoadmapOutputTopic                `json:"topics"`
}

type GetRoadmapOutputProgression struct {
	TotalTopics          int       `json:"total_topics"`
	TotalFinishedTopics  int       `json:"finished_topics"`
	CompletionPercentage float64   `json:"completion_percentage"`
	IsFinished           bool      `json:"is_finished"`
	FinishedAt           time.Time `json:"finished_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type GetRoadmapOutputCreator struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type GetRoadmapOutputTopic struct {
	ID                  int                     `json:"id"`
	RoadmapID           int                     `json:"roadmap_id"`
	ParentID            int                     `json:"parent_id"`
	Title               string                  `json:"title"`
	Slug                string                  `json:"slug"`
	Description         string                  `json:"description"`
	ProTips             string                  `json:"pro_tips"`
	Order               int                     `json:"order"`
	IsFinished          bool                    `json:"is_finished"`
	FinishedAt          time.Time               `json:"finished_at"`
	ExternalSearchQuery string                  `json:"external_search_query"`
	Subtopics           []GetRoadmapOutputTopic `json:"subtopics"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
}

type GetRoadmapOutputPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}

type GetRoadmapOutputRating struct {
	IsRated                        bool      `json:"is_rated"`
	RoadmapID                      int       `json:"roadmap_id"`
	ProgressionTotalTopics         int       `json:"progression_total_topics"`
	ProgressionTotalFinishedTopics int       `json:"progression_total_finished_topics"`
	Rating                         int       `json:"rating"`
	Comment                        string    `json:"comment"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}
