package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) GetUserRoadmapRating(w http.ResponseWriter, r *http.Request) {
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	ctx := r.Context()
	auth := auth.FromContext(ctx)
	output, err := a.application.GetUserRoadmapRating(r.Context(), io.GetUserRoadmapRatingInput{
		Slug:      slug,
		AccountID: auth.AccountID,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Topic Details.", output)
}
