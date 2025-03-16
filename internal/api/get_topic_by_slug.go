package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) GetTopicBySlug(w http.ResponseWriter, r *http.Request) {
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	output, err := a.application.GetTopicBySlug(r.Context(), slug)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Topic Details.", output)
}
