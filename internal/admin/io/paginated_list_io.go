package io

type PaginatedListOutput[T any] struct {
	Total       uint64 `json:"total"`
	TotalPages  uint64 `json:"total_pages"`
	CurrentPage uint64 `json:"current_page"`
	Items       []T    `json:"items"`
}

type PaginatedListInput struct {
	Page  uint64 `json:"page" query:"page"`
	Limit uint64 `json:"limit" query:"limit"`
}
