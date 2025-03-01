package domain

import (
	"errors"
	"time"
)

const (
	PersonalizationOptionsTable = "personalization_options"
)

var (
	ErrPersonalizationOptionsNotFound = errors.New("personalization options not found")
)

type SkillLevel string

const (
	SkillLevelBeginner     SkillLevel = "beginner"
	SkillLevelIntermediate SkillLevel = "intermediate"
	SkillLevelAdvanced     SkillLevel = "advanced"
)

func (s SkillLevel) String() string {
	return string(s)
}

type PersonalizationOptions struct {
	ID                    int
	AccountID             int
	RoadmapID             int
	DailyTimeAvailability time.Duration
	TotalDuration         time.Duration
	SkillLevel            SkillLevel
	AdditionalInfo        string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func NewPersonalizationOptions(accountID, roadmapID int, dailyTimeAvailability, totalDuration time.Duration, skillLevel SkillLevel, additionalInfo string) *PersonalizationOptions {
	return &PersonalizationOptions{
		AccountID:             accountID,
		RoadmapID:             roadmapID,
		DailyTimeAvailability: dailyTimeAvailability,
		TotalDuration:         totalDuration,
		SkillLevel:            skillLevel,
		AdditionalInfo:        additionalInfo,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}

func (p *PersonalizationOptions) IsZero() bool {
	return p.ID == 0 &&
		p.AccountID == 0 &&
		p.RoadmapID == 0 &&
		p.DailyTimeAvailability == 0 &&
		p.TotalDuration == 0 &&
		p.SkillLevel == "" &&
		p.AdditionalInfo == ""
}
