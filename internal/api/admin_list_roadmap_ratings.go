package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/filter"
)

func (a *API) AdminListRoadmapRatings(w http.ResponseWriter, r *http.Request) {
	id, err := a.ParamInt(r, "id")
	if err != nil {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	filters, err := filter.FromRequest(r)
	if err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData.With(err))
		return
	}

	filters.ID = id
	filters.AccountID = 0
	output, err := a.adminApp.ListRoadmapRatings(r.Context(), filters)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap Ratings.", output)
}
