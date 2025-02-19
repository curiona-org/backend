package youtube

import (
	"context"

	"google.golang.org/api/option"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	baseyoutube "google.golang.org/api/youtube/v3"
)

const (
	maxResults = 2
)

type Client interface {
	Search(ctx context.Context, query string) ([]*SearchResult, error)
}

type client struct {
	secret string
	tracer trace.Tracer
}

func New(secret string) Client {
	tracer := otel.Tracer("youtube")
	return &client{
		secret: secret,
		tracer: tracer,
	}
}

type SearchResult struct {
	Title string
	URL   string
}

func (c *client) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	ctx, span := c.tracer.Start(ctx, "(*client).Search", trace.WithAttributes(attribute.String("query", query)))
	defer span.End()

	service, err := baseyoutube.NewService(ctx,
		option.WithAPIKey(c.secret),
		option.WithScopes(baseyoutube.YoutubeReadonlyScope))
	if err != nil {
		return nil, err
	}

	call := service.Search.
		List([]string{"snippet"}).
		Q(query).
		MaxResults(maxResults)

	response, err := call.Do()
	if err != nil {
		return nil, err
	}

	items := response.Items

	videos := make([]*SearchResult, 0, len(items))
	for i, item := range items {
		if item.Id.Kind == "youtube#video" {
			videoURL := "https://www.youtube.com/watch?v=" + item.Id.VideoId
			video := SearchResult{
				Title: item.Snippet.Title,
				URL:   videoURL,
			}
			span.AddEvent("video", trace.WithAttributes(attribute.String("title", video.Title), attribute.String("url", video.URL)))
			videos = append(videos, &video)
		}

		if i >= maxResults {
			break
		}
	}

	return videos, nil
}
