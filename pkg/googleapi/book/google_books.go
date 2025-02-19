package book

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

const (
	googleBooksVolumeAPIUrl = "https://www.googleapis.com/books/v1/volumes?"
)

type Client interface {
	Search(ctx context.Context, query string) ([]*Volume, error)
}

type googleBooksClient struct {
	secret string
	tracer trace.Tracer
}

type googleBooksResult struct {
	Items []struct {
		VolumeInfo googleBooksVolumeInfo `json:"volumeInfo"`
	} `json:"items"`
}

type googleBooksVolumeInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PageCount   int    `json:"pageCount"`
}

func New(secret string) Client {
	tracer := otel.Tracer("book:google_books")
	return &googleBooksClient{
		secret: secret,
		tracer: tracer,
	}
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
		Fields("items(volumeInfo(title,description,pageCount))").
		MaxResults(2)

	result, err := call.Do()
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("count", len(result.Items)))

	items := result.Items

	volumes := make([]*Volume, 0, len(items))
	for _, item := range items {
		volume := Volume{
			Title:       item.VolumeInfo.Title,
			Description: item.VolumeInfo.Description,
			Pages:       int(item.VolumeInfo.PageCount),
		}
		span.AddEvent("book:google_books:search", trace.WithAttributes(attribute.String("title", volume.Title)))
		volumes = append(volumes, &volume)
	}

	return volumes, nil
}
