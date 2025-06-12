package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/filter"
)

func (a *API) ListUserInProgressRoadmaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters, err := filter.FromRequest(r)
	if err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData.With(err))
		return
	}

	output, err := a.application.ListUserInProgressRoadmaps(ctx, filters)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "List User In Progress Roadmaps.", output)
}
