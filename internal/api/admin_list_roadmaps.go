package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/pkg/pagination"
)

func (a *API) AdminListRoadmaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := pagination.FromRequest(r)
	if err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData.Msg(err.Error()))
		return
	}

	output, err := a.adminApp.ListRoadmaps(ctx, io.PaginatedListInput{
		Page:  pagination.CurrentPage,
		Limit: pagination.Limit,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "List Roadmaps.", output)
}
