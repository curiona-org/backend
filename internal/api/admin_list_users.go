package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/pagination"
)

func (a *API) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := pagination.FromRequest(r)
	if err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData.With(err))
		return
	}

	output, err := a.adminApp.ListUsers(ctx, io.PaginatedListInput{
		Page:  pagination.CurrentPage,
		Limit: pagination.Limit,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "List Users.", output)
}
