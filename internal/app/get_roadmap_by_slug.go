package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/interval"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) GetRoadmapBySlug(ctx context.Context, input io.GetRoadmapInput) (io.GetRoadmapOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetRoadmapBySlug)", trace.WithAttributes(attribute.String("slug", input.Slug)))
	defer span.End()

	roadmap, err := app.repository.Roadmap.GetBySlug(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return io.GetRoadmapOutput{}, cerrors.ErrNotFound.Msg("roadmap")
		}
		return io.GetRoadmapOutput{}, err
	}

	if input.AccountID != 0 {
		progressions, err := app.repository.Roadmap.GetTopicProgressions(ctx, input.AccountID, roadmap.ID)
		if err != nil {
			if errors.Is(err, domain.ErrRoadmapNotFound) {
				return io.GetRoadmapOutput{}, cerrors.ErrNotFound.Msg("roadmap")
			}
			return io.GetRoadmapOutput{}, err
		}

		topicMap := make(map[int]bool)
		for _, progression := range progressions {
			topicMap[progression.TopicID] = progression.IsFinished
		}

		for i := range roadmap.Topics {
			if isFinished, ok := topicMap[roadmap.Topics[i].ID]; ok {
				roadmap.Topics[i].IsFinished = isFinished
				if isFinished {
					roadmap.TotalFinishedTopics++
				}
			}
		}
		roadmap.TotalTopics = len(roadmap.Topics)
	}

	output := io.GetRoadmapOutput{
		ID:                   roadmap.ID,
		Title:                roadmap.Title,
		Slug:                 roadmap.Slug,
		Description:          roadmap.Description,
		TotalTopics:          roadmap.TotalTopics,
		TotalFinishedTopics:  roadmap.TotalFinishedTopics,
		CompletionPercentage: roadmap.CompletionPercentage(),
		CreatedAt:            roadmap.CreatedAt,
		UpdatedAt:            roadmap.UpdatedAt,
		Creator: io.GetRoadmapOutputCreator{
			ID:     roadmap.Account.ID,
			Name:   roadmap.Account.Profile.Name,
			Avatar: roadmap.Account.Profile.Avatar,
		},
		PersonalizationOpts: io.GetRoadmapOutputPersonalizationOptions{
			DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
			TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
			SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
			AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
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
