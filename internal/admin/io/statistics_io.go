package io

type StatisticsOutput struct {
	User    StatisticsUser    `json:"user"`
	Roadmap StatisticsRoadmap `json:"roadmap"`
}

type StatisticsUser struct {
	UsersRegisteredCount uint64 `json:"users_registered_count"`
	UsersSuspendedCount  uint64 `json:"users_suspended_count"`
}

type StatisticsRoadmap struct {
	RoadmapsGeneratedCount uint64 `json:"roadmaps_generated_count"`
	RoadmapsDroppedCount   uint64 `json:"roadmaps_dropped_count"`
}
