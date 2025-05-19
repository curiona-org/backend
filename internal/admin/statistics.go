package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
)

func (app *adminApplication) Statistics(ctx context.Context) (io.StatisticsOutput, error) {
	roadmapsCount, err := app.repository.Roadmap.Count(ctx)
	if err != nil {
		return io.StatisticsOutput{}, err
	}

	output := io.StatisticsOutput{
		User: io.StatisticsUser{
			UsersRegisteredCount: 0,
			UsersSuspendedCount:  0,
		},
		Roadmap: io.StatisticsRoadmap{
			RoadmapsGeneratedCount: roadmapsCount,
			RoadmapsDroppedCount:   0,
		},
	}

	return output, nil
}
