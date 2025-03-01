package io

import (
	"time"

	"github.com/curiona-org/backend/pkg/interval"
)

type ListUserRoadmapsOutput struct {
	TotalRoadmaps int                             `json:"total_roadmaps"`
	Roadmaps      []ListUserRoadmapsOutputRoadmap `json:"roadmaps"`
}

type ListUserRoadmapsOutputRoadmap struct {
	ID                   int                                          `json:"id"`
	Title                string                                       `json:"title"`
	Slug                 string                                       `json:"slug"`
	Description          string                                       `json:"description"`
	TotalTopics          int                                          `json:"total_topics"`
	TotalFinishedTopics  int                                          `json:"finished_topics"`
	CompletionPercentage float64                                      `json:"completion_percentage"`
	CreatedAt            time.Time                                    `json:"created_at"`
	UpdatedAt            time.Time                                    `json:"updated_at"`
	PersonalizationOpts  ListUserRoadmapsOutputPersonalizationOptions `json:"personalization_options"`
}

type ListUserRoadmapsOutputPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}
