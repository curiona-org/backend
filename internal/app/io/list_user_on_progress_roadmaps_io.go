package io

import (
	"time"

	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

type ListUserOnProgressRoadmapsInput = filter.Params
type ListUserOnProgressRoadmapsOutput = filter.FilteredList[ListUserOnProgressRoadmapsOutputItem]

type ListUserOnProgressRoadmapsOutputItem struct {
	ID                  int                                                        `json:"id"`
	Title               string                                                     `json:"title"`
	Slug                string                                                     `json:"slug"`
	Description         string                                                     `json:"description"`
	TotalTopics         int                                                        `json:"total_topics"`
	CreatedAt           time.Time                                                  `json:"created_at"`
	UpdatedAt           time.Time                                                  `json:"updated_at"`
	IsBookmarked        bool                                                       `json:"is_bookmarked"`
	Progression         ListUserOnProgressRoadmapsOutputItemProgression            `json:"progression"`
	PersonalizationOpts ListUserOnProgressRoadmapsOutputItemPersonalizationOptions `json:"personalization_options"`
}

type ListUserOnProgressRoadmapsOutputItemProgression struct {
	TotalTopics          int       `json:"total_topics"`
	TotalFinishedTopics  int       `json:"finished_topics"`
	CompletionPercentage float64   `json:"completion_percentage"`
	IsFinished           bool      `json:"is_finished"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ListUserOnProgressRoadmapsOutputItemPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}
