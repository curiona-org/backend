package io

type MarkTopicInput struct {
	Slug      string `json:"slug"`
	AccountID int    `json:"-"`
}
