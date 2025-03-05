package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := a.ParamInt(r, "id")
	if err != nil {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	output, err := a.adminApp.GetUser(r.Context(), id)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "User details.", output)
}
