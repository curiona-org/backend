package admin

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/interval"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	topicMap := make(map[int][]io.GetRoadmapOutputTopic)

	for _, topic := range roadmap.Topics {
		outputTopic := io.GetRoadmapOutputTopic{
			ID:                  topic.ID,
			RoadmapID:           topic.RoadmapID,
			ParentID:            topic.ParentID,
			Title:               topic.Title,
			Slug:                topic.Slug,
			Description:         topic.Description,
			Order:               topic.Order,
			IsFinished:          topic.IsFinished,
			ExternalSearchQuery: topic.ExternalSearchQuery,
			CreatedAt:           topic.CreatedAt,
			UpdatedAt:           topic.UpdatedAt,
		}

		topicMap[topic.ParentID] = append(topicMap[topic.ParentID], outputTopic)
	}

	var buildTopics func(ctx context.Context, parentID int) []io.GetRoadmapOutputTopic
	buildTopics = func(ctx context.Context, parentID int) []io.GetRoadmapOutputTopic {
		traceCtx, buildTopicsSpan := app.tracer.Start(ctx, "buildTopics", trace.WithAttributes(attribute.Int("parentID", parentID)))
		defer buildTopicsSpan.End()

		outputTopics := topicMap[parentID]
		for i := range outputTopics {
			outputTopics[i].Subtopics = buildTopics(traceCtx, outputTopics[i].ID)
		}

		return outputTopics
	}

	output.Topics = buildTopics(ctx, 0)

	return output, nil
}
