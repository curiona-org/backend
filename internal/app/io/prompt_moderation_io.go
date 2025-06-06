package io

type PromptModerationInput struct {
	AccountID int    `json:"-"`
	Prompt    string `json:"prompt" validate:"required,max=150"`
}

type PromptModerationOutput struct {
	Flagged               bool   `json:"flagged"`
	IsMaxGeneratedRoadmap bool   `json:"is_max_generated_roadmap"`
	Reason                string `json:"reason"`
}
