package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) BookmarkRoadmap(w http.ResponseWriter, r *http.Request) {
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	ctx := r.Context()
	auth := auth.FromContext(ctx)
	err := a.application.BookmarkRoadmap(ctx, io.BookmarkRoadmapInput{
		AccountID: auth.AccountID,
		Slug:      slug,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "bookmarked roadmap", nil)
}
