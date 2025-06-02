package api

import (
	"net/http"
)

func (a *API) AdminStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	output, err := a.adminApp.Statistics(ctx)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Statistics.", output)
}
