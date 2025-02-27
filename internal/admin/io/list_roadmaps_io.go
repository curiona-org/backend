package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/domain/object"
)

type ListRoadmapsInput = PaginatedListInput
type ListRoadmapsOutput = PaginatedListOutput[ListRoadmapsOutputItem]

type ListRoadmapsOutputItem struct {
	ID                   int                                          `json:"id"`
	Title                string                                       `json:"title"`
	Slug                 string                                       `json:"slug"`
	Description          string                                       `json:"description"`
	TotalTopics          int                                          `json:"total_topics"`
	CompletionPercentage float64                                      `json:"completion_percentage"`
	CreatedAt            time.Time                                    `json:"created_at"`
	UpdatedAt            time.Time                                    `json:"updated_at"`
	PersonalizationOpts  ListRoadmapsOutputItemPersonalizationOptions `json:"personalization_options"`
	Creator              ListRoadmapsOutputItemUser                   `json:"creator"`
}

type ListRoadmapsOutputItemPersonalizationOptions struct {
	DailyTimeAvailability object.Interval `json:"daily_time_availability"`
	TotalDuration         object.Interval `json:"total_duration"`
	SkillLevel            string          `json:"skill_level"`
	AdditionalInfo        string          `json:"additional_info"`
}

type ListRoadmapsOutputItemUser struct {
	ID          int         `json:"id"`
	Method      auth.Method `json:"method"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Avatar      string      `json:"avatar"`
	IsSuspended bool        `json:"is_suspended"`
	JoinedAt    time.Time   `json:"joined_at"`
}
