package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/interval"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) GetRoadmapBySlug(ctx context.Context, input io.GetRoadmapInput) (io.GetRoadmapOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetRoadmapBySlug)", trace.WithAttributes(attribute.String("slug", input.Slug)))
	defer span.End()

	log := logger.FromContext(ctx)

	roadmap, err := app.repository.Roadmap.GetBySlug(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return io.GetRoadmapOutput{}, cerrors.ErrNotFound.Msg("roadmap")
		}
		return io.GetRoadmapOutput{}, err
	}

	if input.AccountID != 0 {
		progression, err := app.repository.Roadmap.GetRoadmapProgression(ctx, input.AccountID, roadmap.ID)
		if err != nil {
			log.Err(err).Msg("failed to get roadmap progression")
		} else {
			for i := range roadmap.Topics {
				if topicProgress, ok := progression.TopicProgressionMap[roadmap.Topics[i].ID]; ok {
					roadmap.Topics[i].IsFinished = topicProgress.IsFinished
					roadmap.Topics[i].FinishedAt = topicProgress.FinishedAt
				}
			}
		}

		isBookmarked, err := app.repository.Bookmark.RoadmapIsBookmarked(ctx, input.AccountID, roadmap.ID)
		if err != nil {
			log.Err(err).Msg("failed to get roadmap bookmark status")
		} else {
			roadmap.IsBookmarked = isBookmarked
		}

		roadmap.SetProgression(&progression)
	}

	output := io.GetRoadmapOutput{
		ID:           roadmap.ID,
		Title:        roadmap.Title,
		Slug:         roadmap.Slug,
		Description:  roadmap.Description,
		TotalTopics:  roadmap.TotalTopics,
		CreatedAt:    roadmap.CreatedAt,
		UpdatedAt:    roadmap.UpdatedAt,
		IsBookmarked: roadmap.IsBookmarked,
		Progression: io.GetRoadmapOutputProgression{
			TotalTopics:          roadmap.Progression.TotalTopics,
			TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
			IsFinished:           roadmap.Progression.IsFinished,
			FinishedAt:           roadmap.Progression.FinishedAt,
			CompletionPercentage: roadmap.Progression.CompletionPercentage(),
			CreatedAt:            roadmap.Progression.CreatedAt,
			UpdatedAt:            roadmap.Progression.UpdatedAt,
		},
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
			ProTips:             topic.ProTips,
			Order:               topic.Order,
			IsFinished:          topic.IsFinished,
			FinishedAt:          topic.FinishedAt,
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
