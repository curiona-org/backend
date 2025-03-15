package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/pkg/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

type ListCommunityRoadmapsInput = filter.Params
type ListCommunityRoadmapsOutput = filter.FilteredList[ListCommunityRoadmapsOutputItem]

type ListCommunityRoadmapsOutputItem struct {
	ID                   int                                                   `json:"id"`
	Title                string                                                `json:"title"`
	Slug                 string                                                `json:"slug"`
	Description          string                                                `json:"description"`
	TotalTopics          int                                                   `json:"total_topics"`
	TotalFinishedTopics  int                                                   `json:"finished_topics"`
	CompletionPercentage float64                                               `json:"completion_percentage"`
	CreatedAt            time.Time                                             `json:"created_at"`
	UpdatedAt            time.Time                                             `json:"updated_at"`
	PersonalizationOpts  ListCommunityRoadmapsOutputItemPersonalizationOptions `json:"personalization_options"`
	Creator              ListCommunityRoadmapsOutputItemUser                   `json:"creator"`
}

type ListCommunityRoadmapsOutputItemPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}

type ListCommunityRoadmapsOutputItemUser struct {
	ID          int         `json:"id"`
	Method      auth.Method `json:"method"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Avatar      string      `json:"avatar"`
	IsSuspended bool        `json:"is_suspended"`
	JoinedAt    time.Time   `json:"joined_at"`
}
