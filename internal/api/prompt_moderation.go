package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) PromptModeration(w http.ResponseWriter, r *http.Request) {
	var input io.PromptModerationInput
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
	output, err := a.application.PromptModeration(r.Context(), input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Profile details.", output)
}
