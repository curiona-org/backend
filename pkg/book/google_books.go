package book

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/roadmap-thesis/backend/pkg/httpclient"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	googleBooksVolumeAPIUrl = "https://www.googleapis.com/books/v1/volumes?"
)

type googleBooksClient struct {
	client *httpclient.Client
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

func NewGoogleBooksClient() Client {
	tracer := otel.Tracer("book:google_books")
	return &googleBooksClient{
		client: httpclient.New(1000 * time.Second),
		tracer: tracer,
	}
}

func (b *googleBooksClient) Search(ctx context.Context, query string) ([]*Volume, error) {
	traceCtx, span := b.tracer.Start(ctx, "(*googleBooksClient).Search", trace.WithAttributes(attribute.String("query", query)))
	defer span.End()

	q := url.Values{
		"q":      {query},
		"fields": {"items(volumeInfo(title,description,pageCount))"},
	}

	url := googleBooksVolumeAPIUrl + q.Encode()

	req, _ := http.NewRequestWithContext(traceCtx, http.MethodGet, url, nil)

	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result googleBooksResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	items := result.Items

	volumes := make([]*Volume, 0, len(items))
	for _, item := range items {
		volume := Volume{
			Title:       item.VolumeInfo.Title,
			Description: item.VolumeInfo.Description,
			Pages:       item.VolumeInfo.PageCount,
		}
		span.AddEvent("book:google_books:search", trace.WithAttributes(attribute.String("title", volume.Title)))
		volumes = append(volumes, &volume)
	}

	span.AddEvent("book:google_books:search", trace.WithAttributes(attribute.Int("count", len(volumes))))

	return volumes, nil
}
