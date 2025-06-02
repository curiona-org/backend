package io

type StatisticsOutput struct {
	User    StatisticsUser    `json:"user"`
	Roadmap StatisticsRoadmap `json:"roadmap"`
}

type StatisticsUser struct {
	UsersRegisteredCount uint64 `json:"users_registered_count"`
}

type StatisticsRoadmap struct {
	RoadmapsGeneratedCount uint64 `json:"roadmaps_generated_count"`
	RoadmapsOngoingCount   uint64 `json:"roadmaps_ongoing_count"`
	RoadmapsFinishedCount  uint64 `json:"roadmaps_finished_count"`
}
