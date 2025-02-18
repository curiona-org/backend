package io

import "time"

type GetTopicOutput struct {
	ID          int    `json:"id"`
	RoadmapID   int    `json:"roadmap_id"`
	ParentID    int    `json:"parent_id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Finished    bool   `json:"finished"`

	// Subtopics []GetTopicOutput `json:"subtopics"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
