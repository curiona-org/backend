package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
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

	output, err := a.application.PromptModeration(r.Context(), input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Profile details.", output)
}
