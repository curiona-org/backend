package filter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/curiona-org/backend/internal/auth"
)

// FilteredList is a generic structure that represents a list of items with pagination.
type FilteredList[T any] struct {
	Total       uint64 `json:"total"`
	TotalPages  uint64 `json:"total_pages"`
	CurrentPage uint64 `json:"current_page"`
	Items       []T    `json:"items"`
}

type Filters struct {
	ID        int
	AccountID int
	Search    string
	OrderBy   OrderBy
	Paginator Paginator
	Slug      string
	Options   map[string]any
}

func New(params Params, total uint64) Filters {
	paginator := NewOffsetPaginator(params.CurrentPage, params.Limit, total)
	return Filters{
		ID:        params.ID,
		AccountID: params.AccountID,
		Search:    params.Search,
		OrderBy:   params.OrderBy,
		Paginator: paginator,
	}
}

type Params struct {
	// ID is a generic identifier for the resource being filtered.
	ID          int
	AccountID   int
	Search      string
	OrderBy     OrderBy
	CurrentPage uint64
	Limit       uint64
}

type OrderBy string

const (
	OrderByNewest OrderBy = "newest"
	OrderByOldest OrderBy = "oldest"

	QueryKeySearch = "search"
	QueryKeyOrder  = "order_by"
	QueryKeyPage   = "page"
	QueryKeyLimit  = "limit"
)

func FromRequest(r *http.Request) (Params, error) {
	ctx := r.Context()
	search := strings.TrimSpace(r.FormValue(QueryKeySearch))
	order := OrderBy(strings.TrimSpace(r.FormValue(QueryKeyOrder)))

	var accountID int
	auth := auth.FromContext(ctx)
	if auth != nil {
		accountID = auth.AccountID
	}

	page := uint64(1)
	if v := strings.TrimSpace(r.FormValue(QueryKeyPage)); v != "" {
		var err error
		page, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return Params{}, errors.New("invalid query provided")
		}
	}

	limit := DefaultPageSize
	if v := strings.TrimSpace(r.FormValue(QueryKeyLimit)); v != "" {
		var err error
		limit, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return Params{}, errors.New("invalid limit provided")
		}
	}

	// Cap the maximum limit.
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	params := Params{
		AccountID:   accountID,
		Search:      search,
		OrderBy:     order,
		CurrentPage: page,
		Limit:       limit,
	}

	return params, nil
}
