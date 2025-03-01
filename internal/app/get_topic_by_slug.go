package app

import (
	"context"
	"errors"
	"sync"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/internal/worker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

func (app *application) GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error) {
	traceCtx, span := app.tracer.Start(ctx, "(*application.GetTopicBySlug)", trace.WithAttributes(attribute.String("slug", slug)))
	defer span.End()

	topic, err := app.repository.Topic.GetBySlug(traceCtx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) {
			return io.GetTopicOutput{}, cerrors.ErrNotFound.Msg("topic")
		}
		return io.GetTopicOutput{}, err
	}

	output := io.GetTopicOutput{
		ID:          topic.ID,
		AccountID:   topic.AccountID,
		RoadmapID:   topic.RoadmapID,
		ParentID:    topic.ParentID,
		Title:       topic.Title,
		Slug:        topic.Slug,
		Description: topic.Description,
		Order:       topic.Order,
		IsFinished:  topic.IsFinished,
		CreatedAt:   topic.CreatedAt,
		UpdatedAt:   topic.UpdatedAt,
	}

	if !topic.HasResources() {
		var group errgroup.Group
		var mu sync.Mutex

		group.Go(func() error {
			return app.searchYoutubeExternalResources(traceCtx, &mu, &topic)
		})

		group.Go(func() error {
			return app.searchGoogleBooksExternalResources(traceCtx, &mu, &topic)
		})

		if err := group.Wait(); err != nil {
			return io.GetTopicOutput{}, err
		}
	}

	app.mapExternalResourcesOutput(&output, topic)

	return output, nil
}

func (app *application) searchYoutubeExternalResources(ctx context.Context, mu *sync.Mutex, topic *domain.Topic) error {
	youtubeSearchCtx, youtubeSearchSpan := app.tracer.Start(ctx, "(*application.GetTopicBySlug).youtubeSearch")
	defer youtubeSearchSpan.End()

	log := logger.FromContext(youtubeSearchCtx)

	searchResult, err := app.youtube.Search(youtubeSearchCtx, topic.ExternalSearchQuery)
	if err != nil {
		log.Warn().Msg("failed to search youtube")
		err := app.worker.EnqueueSearchYoutubeExternalResources(youtubeSearchCtx,
			worker.SearchYoutubeExternalResourcesInput{
				TopicID:     topic.ID,
				SearchQuery: topic.ExternalSearchQuery,
			})
		if err != nil {
			log.Error().Err(err).Msg("failed to enqueue search youtube external resources")
		}
		log.Info().Msg("enqueued search youtube external resources")
		return nil
	}

	externalResources := make([]*domain.ExternalResource, 0)
	for _, result := range searchResult {
		externalResource := domain.NewExternalResource(
			topic.ID,
			result.Title,
			result.URL,
			domain.ExternalResourceTypeYoutube,
		)

		mu.Lock()
		topic.AddResource(*externalResource)
		mu.Unlock()
		externalResources = append(externalResources, externalResource)
	}

	if len(externalResources) == 0 {
		return nil
	}

	return app.repository.ExternalResource.BulkSave(youtubeSearchCtx, topic.ID, externalResources)
}

func (app *application) searchGoogleBooksExternalResources(ctx context.Context, mu *sync.Mutex, topic *domain.Topic) error {
	bookSearchCtx, bookSearchSpan := app.tracer.Start(ctx, "(*application.GetTopicBySlug).bookSearch")
	defer bookSearchSpan.End()

	log := logger.FromContext(bookSearchCtx)

	searchResult, err := app.googleBooks.Search(bookSearchCtx, topic.ExternalSearchQuery)
	if err != nil {
		log.Warn().Msg("failed to search google books")
		err := app.worker.EnqueueSearchGoogleBooksExternalResources(bookSearchCtx,
			worker.SearchGoogleBooksExternalResourcesInput{
				TopicID:     topic.ID,
				SearchQuery: topic.ExternalSearchQuery,
			})
		if err != nil {
			log.Error().Err(err).Msg("failed to enqueue search google books external resources")
		}
		log.Info().Msg("enqueued search google books external resources")
		return nil
	}

	externalResources := make([]*domain.ExternalResource, 0)
	for _, result := range searchResult {
		externalResource := domain.NewExternalResource(
			topic.ID,
			result.Title,
			"",
			domain.ExternalResourceTypeBook,
		)

		mu.Lock()
		topic.AddResource(*externalResource)
		mu.Unlock()
		externalResources = append(externalResources, externalResource)
	}

	if len(externalResources) == 0 {
		return nil
	}

	return app.repository.ExternalResource.BulkSave(bookSearchCtx, topic.ID, externalResources)
}

func (app *application) mapExternalResourcesOutput(output *io.GetTopicOutput, topic domain.Topic) {
	for _, resource := range topic.GetYoutubeResources() {
		if len(output.ExternalResources.YoutubeVideos) == 0 {
			output.ExternalResources.YoutubeVideos = make([]io.GetTopicOutputExternalResourceItem, 0)
		}

		item := io.GetTopicOutputExternalResourceItem{
			Title: resource.Title,
			URL:   resource.URL,
		}

		output.ExternalResources.YoutubeVideos = append(output.ExternalResources.YoutubeVideos, item)
	}

	for _, resource := range topic.GetBookResources() {
		if len(output.ExternalResources.Books) == 0 {
			output.ExternalResources.Books = make([]io.GetTopicOutputExternalResourceItem, 0)
		}

		item := io.GetTopicOutputExternalResourceItem{
			Title: resource.Title,
			URL:   resource.URL,
		}

		output.ExternalResources.Books = append(output.ExternalResources.Books, item)
	}

	for _, resource := range topic.GetArticleResources() {
		if len(output.ExternalResources.Articles) == 0 {
			output.ExternalResources.Articles = make([]io.GetTopicOutputExternalResourceItem, 0)
		}

		item := io.GetTopicOutputExternalResourceItem{
			Title: resource.Title,
			URL:   resource.URL,
		}

		output.ExternalResources.Articles = append(output.ExternalResources.Articles, item)
	}
}
