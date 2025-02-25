package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/auth"
)

func (a *API) ListUserRoadmaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auth := auth.TokenFromContext(ctx)
	output, err := a.application.ListUserRoadmaps(ctx, auth.AccountID)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "List User Roadmaps.", output)
}
