package app

import (
	"context"
	"errors"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/internal/domain/object"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

func (app *application) GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error) {
	traceCtx, span := app.tracer.Start(ctx, "(*application.GetTopicBySlug)", trace.WithAttributes(attribute.String("slug", slug)))
	defer span.End()

	topic, err := app.repository.Topic().GetBySlug(traceCtx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) {
			return io.GetTopicOutput{}, apperrors.ResourceNotFound("topic")
		}
		return io.GetTopicOutput{}, err
	}

	output := io.GetTopicOutput{
		ID:          topic.ID,
		RoadmapID:   topic.RoadmapID,
		ParentID:    topic.ParentID,
		Title:       topic.Title,
		Slug:        topic.Slug,
		Description: topic.Description,
		Order:       topic.Order,
		Finished:    topic.Finished,
		CreatedAt:   topic.CreatedAt,
		UpdatedAt:   topic.UpdatedAt,
	}

	if !topic.HasResources() {
		var group errgroup.Group

		group.Go(func() error {
			youtubeSearchCtx, youtubeSearchSpan := app.tracer.Start(traceCtx, "(*application.GetTopicBySlug).youtubeSearch")
			defer youtubeSearchSpan.End()
			searchResult, err := app.youtube.Search(youtubeSearchCtx, topic.Title)
			if err != nil {
				return err
			}

			externalResources := make([]*domain.ExternalResource, 0)
			for _, result := range searchResult {
				externalResource := domain.NewExternalResource(
					topic.ID,
					result.Title,
					result.URL,
					object.ExternalResourceTypeYoutube,
				)

				topic.AddResource(*externalResource)
				externalResources = append(externalResources, externalResource)
			}

			return app.repository.ExternalResource().BulkSave(traceCtx, topic.ID, externalResources)
		})

		group.Go(func() error {
			bookSearchCtx, bookSearchSpan := app.tracer.Start(traceCtx, "(*application.GetTopicBySlug).bookSearch")
			defer bookSearchSpan.End()
			searchResult, err := app.googleBooks.Search(bookSearchCtx, topic.Title)
			if err != nil {
				return err
			}

			externalResources := make([]*domain.ExternalResource, 0)
			for _, result := range searchResult {
				externalResource := domain.NewExternalResource(
					topic.ID,
					result.Title,
					"",
					object.ExternalResourceTypeBook,
				)

				topic.AddResource(*externalResource)
				externalResources = append(externalResources, externalResource)
			}

			return app.repository.ExternalResource().BulkSave(traceCtx, topic.ID, externalResources)
		})

		if err := group.Wait(); err != nil {
			return io.GetTopicOutput{}, err
		}
	}

	app.mapExternalResourcesOutput(&output, topic)

	return output, nil
}

func (app *application) mapExternalResourcesOutput(output *io.GetTopicOutput, topic domain.Topic) {
	for _, resource := range topic.GetYoutubeResources() {
		if len(output.ExternalResources.YoutubeVideos) == 0 {
			output.ExternalResources.YoutubeVideos = make([]io.GetTopicOutputExternalResourceItem, 0)
		}

		item := io.GetTopicOutputExternalResourceItem{
			TopicID: resource.TopicID,
			Title:   resource.Title,
			URL:     resource.URL,
		}

		output.ExternalResources.YoutubeVideos = append(output.ExternalResources.YoutubeVideos, item)
	}

	for _, resource := range topic.GetBookResources() {
		if len(output.ExternalResources.Books) == 0 {
			output.ExternalResources.Books = make([]io.GetTopicOutputExternalResourceItem, 0)
		}

		item := io.GetTopicOutputExternalResourceItem{
			TopicID: resource.TopicID,
			Title:   resource.Title,
			URL:     resource.URL,
		}

		output.ExternalResources.Books = append(output.ExternalResources.Books, item)
	}

	for _, resource := range topic.GetArticleResources() {
		if len(output.ExternalResources.Articles) == 0 {
			output.ExternalResources.Articles = make([]io.GetTopicOutputExternalResourceItem, 0)
		}

		item := io.GetTopicOutputExternalResourceItem{
			TopicID: resource.TopicID,
			Title:   resource.Title,
			URL:     resource.URL,
		}

		output.ExternalResources.Articles = append(output.ExternalResources.Articles, item)
	}
}
