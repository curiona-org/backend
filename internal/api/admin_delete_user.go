package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := a.ParamInt(r, "id")
	if err != nil {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	if err := a.adminApp.DeleteUser(r.Context(), id); err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "User deleted.", nil)
}
