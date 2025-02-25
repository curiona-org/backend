package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) DeleteUserRoadmap(w http.ResponseWriter, r *http.Request) {
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	ctx := r.Context()
	auth := auth.TokenFromContext(ctx)
	err := a.application.DeleteUserRoadmap(ctx, io.DeleteUserRoadmapInput{
		AccountID: auth.AccountID,
		Slug:      slug,
	})
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap deleted.", nil)
}
