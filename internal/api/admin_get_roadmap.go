package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) AdminGetRoadmap(w http.ResponseWriter, r *http.Request) {
	id, err := a.ParamInt(r, "id")
	if err != nil {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	output, err := a.adminApp.GetRoadmap(r.Context(), id)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap details.", output)
}
