package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) RegenerateRoadmap(w http.ResponseWriter, r *http.Request) {
	var input io.RegenerateRoadmapInput

	if err := a.Bind(r.Body, &input); err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData)
		return
	}

	if err := a.validator.Validate(&input); err != nil {
		a.handleError(w, r, err)
		return
	}

	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}
	input.Slug = slug

	ctx := r.Context()

	auth := auth.FromContext(ctx)
	input.AccountID = auth.AccountID
	output, err := a.application.RegenerateRoadmap(ctx, input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	if output.Flagged {
		// TODO: should be a different status code, but for now we use 200 OK
		a.render.OK(w, "Flagged content detected", nil)
		return
	}

	a.render.OK(w, "Roadmap generated successfully", output)
}
