package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) GenerateRoadmap(w http.ResponseWriter, r *http.Request) {
	var input io.GenerateRoadmapInput

	if err := a.Bind(r.Body, &input); err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData)
		return
	}

	if err := a.validator.Validate(&input); err != nil {
		a.handleError(w, r, err)
		return
	}

	ctx := r.Context()

	auth := auth.FromContext(ctx)
	input.AccountID = auth.AccountID
	output, err := a.application.GenerateRoadmap(ctx, input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	if output.Flagged {
		// TODO: should be a different status code, but for now we use 200 OK
		a.render.OK(w, "Flagged content detected", output)
		return
	}

	a.render.OK(w, "Roadmap generated successfully", output)
}
