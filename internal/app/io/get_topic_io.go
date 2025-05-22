package io

import "time"

type GetTopicOutput struct {
	ID          int       `json:"id"`
	AccountID   int       `json:"account_id"`
	RoadmapID   int       `json:"roadmap_id"`
	ParentID    int       `json:"parent_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ProTips     string    `json:"pro_tips"`
	Order       int       `json:"order"`
	IsFinished  bool      `json:"is_finished"`
	FinishedAt  time.Time `json:"finished_at"`

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
	Title    string `json:"title"`
	Author   string `json:"author"`
	URL      string `json:"url"`
	CoverURL string `json:"cover_url"`
	Length   string `json:"length"`
}
