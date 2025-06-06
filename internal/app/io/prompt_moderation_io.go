package io

type PromptModerationInput struct {
	Prompt string `json:"prompt" validate:"required,max=150"`
}

type PromptModerationOutput struct {
	Flagged bool   `json:"flagged"`
	Reason  string `json:"reason"`
}
