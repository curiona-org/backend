package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) GetRoadmapBySlug(w http.ResponseWriter, r *http.Request) {
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	ctx := r.Context()
	input := io.GetRoadmapInput{
		Slug: slug,
	}

	auth := auth.FromContext(ctx)
	if auth != nil {
		input.AccountID = auth.AccountID
	}

	output, err := a.application.GetRoadmapBySlug(r.Context(), input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap details.", output)
}
