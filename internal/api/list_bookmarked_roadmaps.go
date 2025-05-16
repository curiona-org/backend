package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/filter"
)

func (a *API) ListBookmarkedRoadmaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters, err := filter.FromRequest(r)
	if err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData.With(err))
		return
	}

	auth := auth.FromContext(ctx)
	output, err := a.application.ListBookmarkedRoadmaps(ctx, io.ListBookmarkedRoadmapsInput{
		Params:    filters,
		AccountID: auth.AccountID,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "List Bookmarked Roadmaps.", output)
}
