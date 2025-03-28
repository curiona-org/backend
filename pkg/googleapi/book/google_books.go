package book

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

const (
	maxResults = 2
)

type Client interface {
	Search(ctx context.Context, query string) ([]*Volume, error)
}

type googleBooksClient struct {
	secret string
	tracer trace.Tracer
}

func New(secret string) Client {
	tracer := otel.Tracer("book:google_books")
	return &googleBooksClient{
		secret: secret,
		tracer: tracer,
	}
}

type Volume struct {
	Title   string
	URL     string
	Authors string
	Cover   string
	Pages   int64
}

func (b *googleBooksClient) Search(ctx context.Context, query string) ([]*Volume, error) {
	traceCtx, span := b.tracer.Start(ctx, "(*googleBooksClient).Search", trace.WithAttributes(attribute.String("query", query)))
	defer span.End()

	service, err := books.NewService(traceCtx,
		option.WithAPIKey(b.secret),
		option.WithScopes(books.BooksScope))
	if err != nil {
		return nil, err
	}

	call := service.Volumes.
		List(query).
		Fields("items(volumeInfo(title,authors,imageLinks(thumbnail),canonicalVolumeLink,pageCount))").
		MaxResults(maxResults)

	result, err := call.Do()
	if err != nil {
		span.SetStatus(codes.Error, "failed to fetch google books")
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("count", len(result.Items)))

	items := result.Items

	volumes := make([]*Volume, 0, len(items))
	for i, item := range items {
		// To make sure we don't exceed the maxResults if the API returns more than expected
		if i >= maxResults {
			break
		}

		volume := Volume{
			Title:   item.VolumeInfo.Title,
			Authors: strings.Join(item.VolumeInfo.Authors, ", "),
			Cover:   item.VolumeInfo.ImageLinks.Thumbnail,
			URL:     item.VolumeInfo.CanonicalVolumeLink,
			Pages:   item.VolumeInfo.PageCount,
		}

		span.AddEvent("book:google_books:search", trace.WithAttributes(
			attribute.String("title", volume.Title),
			attribute.String("authors", volume.Authors),
			attribute.String("url", volume.URL),
			attribute.String("cover", volume.Cover),
			attribute.Int64("pages", volume.Pages)))

		volumes = append(volumes, &volume)
	}

	return volumes, nil
}
