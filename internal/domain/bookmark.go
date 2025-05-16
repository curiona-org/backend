package domain

import (
	"errors"
	"time"
)

const (
	BookmarkTable = "bookmarks"
)

var (
	ErrBookmarkNotFound = errors.New("bookmark not found")
)

type Bookmark struct {
	AccountID int
	RoadmapID int
	CreatedAt time.Time
}
