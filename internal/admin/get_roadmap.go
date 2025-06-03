package admin

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *adminApplication) GetRoadmap(ctx context.Context, roadmapID int) (io.GetRoadmapOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.GetRoadmap)")
	defer span.End()

	roadmap, err := app.repository.Roadmap.GetByID(ctx, roadmapID)
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return io.GetRoadmapOutput{}, cerrors.ErrNotFound.Msg("roadmap")
		}
		return io.GetRoadmapOutput{}, err
	}

	output := io.GetRoadmapOutput{
		ID:          roadmap.ID,
		Title:       roadmap.Title,
		Description: roadmap.Description,
		Slug:        roadmap.Slug,
		TotalTopics: roadmap.TotalTopics,
		CreatedAt:   roadmap.CreatedAt,
		UpdatedAt:   roadmap.UpdatedAt,
		PersonalizationOpts: io.GetRoadmapOutputPersonalizationOptions{
			DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
			TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
			SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
			AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
		},
		Creator: io.GetRoadmapOutputCreator{
			ID:          roadmap.Account.ID,
			Method:      roadmap.Account.Method,
			Email:       roadmap.Account.Email,
			Name:        roadmap.Account.Profile.Name,
			Avatar:      roadmap.Account.Profile.Avatar,
			IsSuspended: roadmap.Account.IsSuspended,
			JoinedAt:    roadmap.Account.CreatedAt,
		},
	}

	for _, topic := range roadmap.Topics {
		outputTopic := io.GetRoadmapOutputTopic{
			ID:                  topic.ID,
			RoadmapID:           topic.RoadmapID,
			ParentID:            topic.ParentID,
			Title:               topic.Title,
			Slug:                topic.Slug,
			Description:         topic.Description,
			ProTips:             topic.ProTips,
			Order:               topic.Order,
			IsFinished:          topic.IsFinished,
			FinishedAt:          topic.FinishedAt,
			ExternalSearchQuery: topic.ExternalSearchQuery,
			CreatedAt:           topic.CreatedAt,
			UpdatedAt:           topic.UpdatedAt,
		}

		for _, subtopic := range topic.Subtopics {
			outputSubtopic := io.GetRoadmapOutputTopic{
				ID:                  subtopic.ID,
				RoadmapID:           subtopic.RoadmapID,
				ParentID:            subtopic.ParentID,
				Title:               subtopic.Title,
				Slug:                subtopic.Slug,
				Description:         subtopic.Description,
				ProTips:             subtopic.ProTips,
				Order:               subtopic.Order,
				IsFinished:          subtopic.IsFinished,
				FinishedAt:          subtopic.FinishedAt,
				ExternalSearchQuery: subtopic.ExternalSearchQuery,
				CreatedAt:           subtopic.CreatedAt,
				UpdatedAt:           subtopic.UpdatedAt,
			}

			outputTopic.Subtopics = append(outputTopic.Subtopics, outputSubtopic)
		}
		output.Topics = append(output.Topics, outputTopic)
	}

	return output, nil
}
