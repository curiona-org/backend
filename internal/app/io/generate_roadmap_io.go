package io

import (
	"github.com/curiona-org/backend/pkg/interval"
)

type GenerateRoadmapInput struct {
	AccountID              int    `json:"-"`
	Topic                  string `json:"topic" validate:"required"`
	PersonalizationOptions struct {
		DailyTimeAvailability struct {
			Value int           `json:"value" validate:"required"`
			Unit  interval.Unit `json:"unit" validate:"required,oneof=hours minutes"`
		} `json:"daily_time_availability" validate:"required"`
		TotalDuration struct {
			Value int           `json:"value" validate:"required"`
			Unit  interval.Unit `json:"unit" validate:"required,oneof=months weeks days"`
		} `json:"total_duration" validate:"required"`
		SkillLevel     string `json:"skill_level" validate:"required,oneof=beginner intermediate advanced"`
		AdditionalInfo string `json:"additional_info" validate:"omitempty"`
	} `json:"personalization_options" validate:"required"`
}

type GenerateRoadmapOutput struct {
	Slug string `json:"slug"`
}
