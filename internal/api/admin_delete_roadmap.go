package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) AdminDeleteRoadmap(w http.ResponseWriter, r *http.Request) {
	id, err := a.ParamInt(r, "id")
	if err != nil {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	if err := a.adminApp.DeleteRoadmap(r.Context(), id); err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap deleted.", nil)
}
