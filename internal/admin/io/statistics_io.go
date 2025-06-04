package io

import "time"

type StatisticsOutput struct {
	User    StatisticsUser    `json:"user"`
	Roadmap StatisticsRoadmap `json:"roadmap"`
}

type StatisticsUser struct {
	UsersRegisteredCount uint64 `json:"users_registered_count"`
}

type StatisticsRoadmap struct {
	RoadmapsGeneratedCount      uint64 `json:"roadmaps_generated_count"`
	RoadmapsGeneratedTodayCount uint64 `json:"roadmaps_generated_today_count"`
	RoadmapsOngoingCount        uint64 `json:"roadmaps_ongoing_count"`
	RoadmapsFinishedCount       uint64 `json:"roadmaps_finished_count"`
	HighestRatedRoadmapOutput   `json:"highest_rated_roadmap"`
	MostBookmarkedRoadmapOutput `json:"most_bookmarked_roadmap"`
	MostActiveRoadmapOutput     `json:"most_active_roadmap"`
}

type HighestRatedRoadmapOutput struct {
	ID          int       `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Rating      float64   `json:"rating"`
	RatingCount uint64    `json:"rating_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MostBookmarkedRoadmapOutput struct {
	ID            int       `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	BookmarkCount uint64    `json:"bookmark_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MostActiveRoadmapOutput struct {
	ID            int       `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ActivityCount uint64    `json:"activity_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
