package io

import (
	"time"

	"github.com/roadmap-thesis/backend/internal/domain/object"
)

type GetRoadmapOutput struct {
	ID                   int                                    `json:"id"`
	Title                string                                 `json:"title"`
	Slug                 string                                 `json:"slug"`
	Description          string                                 `json:"description"`
	Creator              GetRoadmapOutputCreator                `json:"creator"`
	Topics               []GetRoadmapOutputTopics               `json:"topics"`
	TotalTopics          int                                    `json:"total_topics"`
	PersonalizationOpts  GetRoadmapOutputPersonalizationOptions `json:"personalization_options"`
	CompletionPercentage float64                                `json:"completion_percentage"`
	CreatedAt            time.Time                              `json:"created_at"`
	UpdatedAt            time.Time                              `json:"updated_at"`
}
type GetRoadmapOutputCreator struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type GetRoadmapOutputTopics struct {
	ID          int                      `json:"id"`
	RoadmapID   int                      `json:"roadmap_id"`
	ParentID    int                      `json:"parent_id"`
	Title       string                   `json:"title"`
	Slug        string                   `json:"slug"`
	Description string                   `json:"description"`
	Order       int                      `json:"order"`
	Finished    bool                     `json:"finished"`
	Subtopics   []GetRoadmapOutputTopics `json:"subtopics"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

type GetRoadmapOutputPersonalizationOptions struct {
	DailyTimeAvailability object.Interval `json:"daily_time_availability"`
	TotalDuration         object.Interval `json:"total_duration"`
	SkillLevel            string          `json:"skill_level"`
	AdditionalInfo        string          `json:"additional_info"`
}
