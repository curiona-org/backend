package io

import "time"

type GetTopicOutput struct {
	ID          int    `json:"id"`
	AccountID   int    `json:"account_id"`
	RoadmapID   int    `json:"roadmap_id"`
	ParentID    int    `json:"parent_id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Finished    bool   `json:"finished"`

	ExternalResources GetTopicOutputExternalResource `json:"external_resources"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetTopicOutputExternalResource struct {
	Books         []GetTopicOutputExternalResourceItem `json:"books"`
	YoutubeVideos []GetTopicOutputExternalResourceItem `json:"youtube_videos"`
	Articles      []GetTopicOutputExternalResourceItem `json:"articles"`
}

type GetTopicOutputExternalResourceItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}
