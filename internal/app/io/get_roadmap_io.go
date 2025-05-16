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
	ID                   int                                    `json:"id"`
	Title                string                                 `json:"title"`
	Slug                 string                                 `json:"slug"`
	Description          string                                 `json:"description"`
	TotalTopics          int                                    `json:"total_topics"`
	TotalFinishedTopics  int                                    `json:"finished_topics"`
	CompletionPercentage float64                                `json:"completion_percentage"`
	CreatedAt            time.Time                              `json:"created_at"`
	UpdatedAt            time.Time                              `json:"updated_at"`
	Creator              GetRoadmapOutputCreator                `json:"creator"`
	PersonalizationOpts  GetRoadmapOutputPersonalizationOptions `json:"personalization_options"`
	Topics               []GetRoadmapOutputTopic                `json:"topics"`
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
	Order               int                     `json:"order"`
	IsFinished          bool                    `json:"is_finished"`
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
