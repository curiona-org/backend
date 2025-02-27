package pagination

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/curiona-org/backend/internal/cerrors"
)

const (
	QueryKeyPage    = "page"
	QueryKeyLimit   = "limit"
	DefaultPageSize = uint64(10)
	MaxPageSize     = uint64(100)
)

type Paginator struct {
	CurrentPage uint64
	Total       uint64
	TotalPages  uint64
	Limit       uint64
	Skip        uint64
}

func NewOffsetPaginator(page, pageSize, total uint64) Paginator {
	var paginator Paginator
	if pageSize > 0 || pageSize < MaxPageSize {
		paginator.Limit = pageSize
	} else {
		paginator.Limit = DefaultPageSize
	}

	totalPages := uint64(math.Ceil(float64(total) / float64(paginator.Limit)))
	paginator.Total = total
	paginator.TotalPages = totalPages
	if page < 1 || totalPages == 0 {
		page = 1
	}

	paginator.CurrentPage = page
	paginator.Skip = (page - 1) * paginator.Limit
	return paginator
}

type Params struct {
	CurrentPage uint64
	Limit       uint64
}

func FromRequest(r *http.Request) (Params, error) {
	page := uint64(1)
	if v := strings.TrimSpace(r.FormValue(QueryKeyPage)); v != "" {
		var err error
		page, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return Params{}, cerrors.ErrInvalidData.Msg("invalid query provided")
		}
	}

	limit := DefaultPageSize
	if v := strings.TrimSpace(r.FormValue(QueryKeyLimit)); v != "" {
		var err error
		limit, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return Params{}, cerrors.ErrInvalidData.Msg("invalid limit provided")
		}
	}

	// Cap the maximum limit.
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	params := Params{
		CurrentPage: page,
		Limit:       limit,
	}

	return params, nil
}
