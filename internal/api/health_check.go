package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	repoOK := a.application.HealthCheck(r.Context())
	if !repoOK {
		a.handleError(w, r, cerrors.ErrInternal.Msg("repository is not healthy"))
		return
	}

	a.render.OK(w, "OK", nil)
}
