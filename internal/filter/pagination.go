package filter

import (
	"math"
)

const (
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
