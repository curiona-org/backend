package io

import (
	"time"

	"github.com/curiona-org/backend/pkg/interval"
)

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
	Topics               []GetRoadmapOutputTopics               `json:"topics"`
	PersonalizationOpts  GetRoadmapOutputPersonalizationOptions `json:"personalization_options"`
}
type GetRoadmapOutputCreator struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type GetRoadmapOutputTopics struct {
	ID                  int                      `json:"id"`
	RoadmapID           int                      `json:"roadmap_id"`
	ParentID            int                      `json:"parent_id"`
	Title               string                   `json:"title"`
	Slug                string                   `json:"slug"`
	Description         string                   `json:"description"`
	Order               int                      `json:"order"`
	Finished            bool                     `json:"finished"`
	ExternalSearchQuery string                   `json:"external_search_query"`
	Subtopics           []GetRoadmapOutputTopics `json:"subtopics"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
}

type GetRoadmapOutputPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}
