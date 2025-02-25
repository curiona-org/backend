package api

import "net/http"

func (a *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	a.render.OK(w, "OK", nil)
}
