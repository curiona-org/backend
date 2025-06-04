package app

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/internal/worker"
	"github.com/sosodev/duration"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

func (app *application) GetTopicBySlug(ctx context.Context, input io.GetTopicInput) (io.GetTopicOutput, error) {
	traceCtx, span := app.tracer.Start(ctx, "(*application.GetTopicBySlug)", trace.WithAttributes(attribute.String("slug", input.Slug)))
	defer span.End()

	topic, err := app.repository.Topic.GetBySlug(traceCtx, filter.Filters{
		AccountID: input.AccountID,
		Slug:      input.Slug,
	})
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
		ProTips:     topic.ProTips,
		Order:       topic.Order,
		IsFinished:  topic.IsFinished,
		FinishedAt:  topic.FinishedAt,
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

	app.mapExternalResourcesOutput(traceCtx, topic.Resources, &output)

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
			result.Channel,
			result.URL,
			result.Thumbnail,
			result.Duration,
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
			result.Authors,
			result.URL,
			result.Cover,
			strconv.FormatInt(result.Pages, 10),
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

func (app *application) mapExternalResourcesOutput(ctx context.Context, resources []domain.ExternalResource, output *io.GetTopicOutput) {
	output.ExternalResources.YoutubeVideos = make([]io.GetTopicOutputExternalResourceItem, 0)
	output.ExternalResources.Books = make([]io.GetTopicOutputExternalResourceItem, 0)
	output.ExternalResources.Articles = make([]io.GetTopicOutputExternalResourceItem, 0)

	log := logger.FromContext(ctx)
	for _, resource := range resources {
		contentLength := resource.Length

		if resource.IsYoutube() {
			// Parse ISO 8601 duration
			parsedContentLength, err := duration.Parse(resource.Length)
			if err != nil {
				log.Error().Err(err).Msg("failed to parse youtube video duration")
				contentLength = resource.Length
			} else {
				contentLength = parsedContentLength.ToTimeDuration().String()
			}
		}

		item := io.GetTopicOutputExternalResourceItem{
			Title:    resource.Title,
			Author:   resource.Author,
			URL:      resource.URL,
			CoverURL: resource.CoverURL,
			Length:   contentLength,
		}

		switch {
		case resource.IsYoutube():
			output.ExternalResources.YoutubeVideos = append(output.ExternalResources.YoutubeVideos, item)
		case resource.IsBook():
			output.ExternalResources.Books = append(output.ExternalResources.Books, item)
		case resource.IsArticle():
			output.ExternalResources.Articles = append(output.ExternalResources.Articles, item)
		}
	}
}
