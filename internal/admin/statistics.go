package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/domain"
	"golang.org/x/sync/errgroup"
)

func (app *adminApplication) Statistics(ctx context.Context) (io.StatisticsOutput, error) {
	accountsCount, err := app.repository.Account.Count(ctx)
	if err != nil {
		return io.StatisticsOutput{}, err
	}

	var group errgroup.Group
	var roadmapsCount uint64
	group.Go(func() error {
		var err error
		roadmapsCount, err = app.repository.Roadmap.Count(ctx)
		return err
	})

	var roadmapsInProgressCount uint64
	group.Go(func() error {
		var err error
		roadmapsInProgressCount, err = app.repository.Roadmap.CountInProgress(ctx)
		return err
	})

	var roadmapsFinishedCount uint64
	group.Go(func() error {
		var err error
		roadmapsFinishedCount, err = app.repository.Roadmap.CountFinished(ctx)
		return err
	})

	var roadmapsGeneratedTodayCount uint64
	group.Go(func() error {
		roadmapsGeneratedTodayCount, err = app.repository.Roadmap.CountGeneratedToday(ctx)
		return err
	})

	var highestRatedRoadmap domain.Roadmap
	group.Go(func() error {
		highestRatedRoadmap, err = app.repository.Roadmap.GetHighestRated(ctx)
		return err
	})

	var mostBookmarkedRoadmap domain.Roadmap
	group.Go(func() error {
		mostBookmarkedRoadmap, err = app.repository.Roadmap.GetMostBookmarked(ctx)
		return err
	})

	var mostActiveRoadmap domain.Roadmap
	group.Go(func() error {
		mostActiveRoadmap, err = app.repository.Roadmap.GetMostActive(ctx)
		return err
	})

	if err := group.Wait(); err != nil {
		return io.StatisticsOutput{}, err
	}

	output := io.StatisticsOutput{
		User: io.StatisticsUser{
			UsersRegisteredCount: accountsCount,
		},
		Roadmap: io.StatisticsRoadmap{
			RoadmapsGeneratedCount:      roadmapsCount,
			RoadmapsGeneratedTodayCount: roadmapsGeneratedTodayCount,
			RoadmapsOngoingCount:        roadmapsInProgressCount,
			RoadmapsFinishedCount:       roadmapsFinishedCount,

			HighestRatedRoadmapOutput: io.HighestRatedRoadmapOutput{
				ID:          highestRatedRoadmap.ID,
				Slug:        highestRatedRoadmap.Slug,
				Title:       highestRatedRoadmap.Title,
				Description: highestRatedRoadmap.Description,
				Rating:      highestRatedRoadmap.AverageRating,
				RatingCount: highestRatedRoadmap.TotalRatings,
				CreatedAt:   highestRatedRoadmap.CreatedAt,
				UpdatedAt:   highestRatedRoadmap.UpdatedAt,
			},

			MostBookmarkedRoadmapOutput: io.MostBookmarkedRoadmapOutput{
				ID:            mostBookmarkedRoadmap.ID,
				Slug:          mostBookmarkedRoadmap.Slug,
				Title:         mostBookmarkedRoadmap.Title,
				Description:   mostBookmarkedRoadmap.Description,
				BookmarkCount: uint64(mostBookmarkedRoadmap.TotalBookmarks),
				CreatedAt:     mostBookmarkedRoadmap.CreatedAt,
				UpdatedAt:     mostBookmarkedRoadmap.UpdatedAt,
			},

			MostActiveRoadmapOutput: io.MostActiveRoadmapOutput{
				ID:            mostActiveRoadmap.ID,
				Slug:          mostActiveRoadmap.Slug,
				Title:         mostActiveRoadmap.Title,
				Description:   mostActiveRoadmap.Description,
				ActivityCount: mostActiveRoadmap.TotalActive,
				CreatedAt:     mostActiveRoadmap.CreatedAt,
				UpdatedAt:     mostActiveRoadmap.UpdatedAt,
			},
		},
	}

	return output, nil
}
