package io

type RateRoadmapInput struct {
	AccountID int    `json:"-"`
	Slug      string `json:"slug"`

	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"omitempty,max=500"`
}
