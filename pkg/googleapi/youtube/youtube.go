package youtube

import (
	"context"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	secrets           []string
	nextSecretCounter uint32
	tracer            trace.Tracer
}

func New(secrets []string) Client {
	tracer := otel.Tracer("youtube")
	return &client{
		secrets: secrets,
		tracer:  tracer,
	}
}

type SearchResult struct {
	id        string
	Title     string
	URL       string
	Channel   string
	Thumbnail string
	Duration  string
}

func (c *client) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	ctx, span := c.tracer.Start(ctx, "(*client).Search", trace.WithAttributes(attribute.String("query", query)))
	defer span.End()

	service, err := baseyoutube.NewService(ctx,
		option.WithAPIKey(c.nextSecret()),
		option.WithScopes(baseyoutube.YoutubeReadonlyScope))
	if err != nil {
		return nil, err
	}

	searchCall := service.Search.
		List([]string{"snippet"}).
		Q(query).
		MaxResults(maxResults)

	searchResponse, err := searchCall.Do()
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch youtube videos")
		span.RecordError(err)
		return nil, err
	}

	items := searchResponse.Items

	videoIDs := make([]string, 0, len(items))
	videos := make([]*SearchResult, 0, len(items))
	videoIDMap := make(map[string]*SearchResult)
	for i, item := range items {
		// To make sure we don't exceed the maxResults if the API returns more than expected
		if i >= maxResults {
			break
		}

		if item.Id.Kind == "youtube#video" {
			var thumbnail string
			if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.High != nil {
				thumbnail = item.Snippet.Thumbnails.High.Url
			} else {
				thumbnail = "https://placehold.co/400x200?text=No+Thumbnail"
			}

			videoURL := "https://www.youtube.com/watch?v=" + item.Id.VideoId
			video := SearchResult{
				Title:     item.Snippet.Title,
				URL:       videoURL,
				Channel:   item.Snippet.ChannelTitle,
				Thumbnail: thumbnail,
				Duration:  item.Id.VideoId,
			}
			span.AddEvent("video", trace.WithAttributes(
				attribute.String("title", video.Title),
				attribute.String("channel", video.Channel),
				attribute.String("url", video.URL)))
			videos = append(videos, &video)
			videoIDs = append(videoIDs, item.Id.VideoId)
			videoIDMap[item.Id.VideoId] = &video
		}
	}

	videoCall := service.Videos.List([]string{"contentDetails"}).Id(videoIDs...)
	videoResponse, err := videoCall.Do()
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch youtube video details")
		span.RecordError(err)
		return nil, err
	}

	for _, item := range videoResponse.Items {
		videoID := item.Id
		video, ok := videoIDMap[videoID]
		if !ok {
			span.SetStatus(codes.Error, "video not found in map")
			span.RecordError(err)
			continue
		}

		duration := item.ContentDetails.Duration
		video.Duration = duration
		span.AddEvent("video duration", trace.WithAttributes(
			attribute.String("video_id", videoID),
			attribute.String("duration", duration)))
	}

	return videos, nil
}

func (c *client) nextSecret() string {
	if len(c.secrets) == 0 {
		return ""
	}

	n := atomic.AddUint32(&c.nextSecretCounter, 1)
	key := c.secrets[(int(n)-1)%len(c.secrets)]

	log.Debug().Msgf("Using YouTube API key: %s", key)
	return key
}
