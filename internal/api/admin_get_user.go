package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/filter"
)

func (a *API) AdminGetUser(w http.ResponseWriter, r *http.Request) {
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

	filters.AccountID = id

	output, err := a.adminApp.GetUser(r.Context(), filters)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "User details.", output)
}
