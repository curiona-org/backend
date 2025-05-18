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
	auth := auth.FromContext(ctx)
	output, err := a.application.GetRoadmapBySlug(r.Context(), io.GetRoadmapInput{
		AccountID: auth.AccountID,
		Slug:      slug,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap details.", output)
}
