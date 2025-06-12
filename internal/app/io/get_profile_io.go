package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
)

type GetProfileOutput struct {
	ID         int                        `json:"id"`
	Method     auth.Method                `json:"method"`
	Email      string                     `json:"email"`
	Name       string                     `json:"name"`
	Avatar     string                     `json:"avatar"`
	JoinedAt   time.Time                  `json:"joined_at"`
	Statistics GetProfileOutputStatistics `json:"statistics"`
}

type GetProfileOutputStatistics struct {
	TotalGeneratedRoadmaps  uint64 `json:"total_generated_roadmaps"`
	TotalInProgressRoadmaps uint64 `json:"total_in_progress_roadmaps"`
	TotalFinishedRoadmaps   uint64 `json:"total_finished_roadmaps"`
	TotalBookmarkedRoadmaps uint64 `json:"total_bookmarked_roadmaps"`
}
