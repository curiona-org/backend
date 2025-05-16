package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/pkg/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

type ListBookmarkedRoadmapsInput struct {
	filter.Params
	AccountID int
}
type ListBookmarkedRoadmapsOutput = filter.FilteredList[ListBookmarkedRoadmapsOutputItem]

type ListBookmarkedRoadmapsOutputItem struct {
	ID                   int                                                    `json:"id"`
	Title                string                                                 `json:"title"`
	Slug                 string                                                 `json:"slug"`
	Description          string                                                 `json:"description"`
	TotalTopics          int                                                    `json:"total_topics"`
	TotalFinishedTopics  int                                                    `json:"finished_topics"`
	CompletionPercentage float64                                                `json:"completion_percentage"`
	CreatedAt            time.Time                                              `json:"created_at"`
	UpdatedAt            time.Time                                              `json:"updated_at"`
	PersonalizationOpts  ListBookmarkedRoadmapsOutputItemPersonalizationOptions `json:"personalization_options"`
	Creator              ListBookmarkedRoadmapsOutputItemUser                   `json:"creator"`
}

type ListBookmarkedRoadmapsOutputItemPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}

type ListBookmarkedRoadmapsOutputItemUser struct {
	ID          int         `json:"id"`
	Method      auth.Method `json:"method"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Avatar      string      `json:"avatar"`
	IsSuspended bool        `json:"is_suspended"`
	JoinedAt    time.Time   `json:"joined_at"`
}
