package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
)

func (app *adminApplication) Statistics(ctx context.Context) (io.StatisticsOutput, error) {
	accountsCount, err := app.repository.Account.Count(ctx)
	if err != nil {
		return io.StatisticsOutput{}, err
	}

	roadmapsCount, err := app.repository.Roadmap.Count(ctx)
	if err != nil {
		return io.StatisticsOutput{}, err
	}

	roadmapsOnProgressCount, err := app.repository.Roadmap.CountOnProgress(ctx)
	if err != nil {
		return io.StatisticsOutput{}, err
	}

	roadmapsFinishedCount, err := app.repository.Roadmap.CountFinished(ctx)
	if err != nil {
		return io.StatisticsOutput{}, err
	}

	output := io.StatisticsOutput{
		User: io.StatisticsUser{
			UsersRegisteredCount: accountsCount,
		},
		Roadmap: io.StatisticsRoadmap{
			RoadmapsGeneratedCount: roadmapsCount,
			RoadmapsOngoingCount:   roadmapsOnProgressCount,
			RoadmapsFinishedCount:  roadmapsFinishedCount,
		},
	}

	return output, nil
}
