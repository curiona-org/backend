package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

type GetUserInput = filter.Params
type GetUserOutput struct {
	ID          int                    `json:"id"`
	Method      auth.Method            `json:"method"`
	Email       string                 `json:"email"`
	Name        string                 `json:"name"`
	Avatar      string                 `json:"avatar"`
	IsSuspended bool                   `json:"is_suspended"`
	IsAdmin     bool                   `json:"is_admin"`
	JoinedAt    time.Time              `json:"joined_at"`
	Roadmaps    ListUserRoadmapsOutput `json:"roadmaps"`
}

type ListUserRoadmapsOutput = filter.FilteredList[ListUserRoadmapsOutputItem]

type ListUserRoadmapsOutputItem struct {
	ID                  int                                              `json:"id"`
	Title               string                                           `json:"title"`
	Slug                string                                           `json:"slug"`
	Description         string                                           `json:"description"`
	TotalTopics         int                                              `json:"total_topics"`
	CreatedAt           time.Time                                        `json:"created_at"`
	UpdatedAt           time.Time                                        `json:"updated_at"`
	IsBookmarked        bool                                             `json:"is_bookmarked"`
	Progression         ListUserRoadmapsOutputItemProgression            `json:"progression"`
	PersonalizationOpts ListUserRoadmapsOutputItemPersonalizationOptions `json:"personalization_options"`
}

type ListUserRoadmapsOutputItemProgression struct {
	TotalTopics          int       `json:"total_topics"`
	TotalFinishedTopics  int       `json:"finished_topics"`
	CompletionPercentage float64   `json:"completion_percentage"`
	IsFinished           bool      `json:"is_finished"`
	FinishedAt           time.Time `json:"finished_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ListUserRoadmapsOutputItemPersonalizationOptions struct {
	DailyTimeAvailability interval.Interval `json:"daily_time_availability"`
	TotalDuration         interval.Interval `json:"total_duration"`
	SkillLevel            string            `json:"skill_level"`
	AdditionalInfo        string            `json:"additional_info"`
}
