package api

import (
	"net/http"
	"strings"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/logger"
)

func (a *API) GetRoadmapBySlug(w http.ResponseWriter, r *http.Request) {
	slug := a.Param(r, "slug")
	if slug == "" {
		a.handleError(w, r, cerrors.ErrNotFound)
		return
	}

	input := &io.GetRoadmapInput{
		Slug: slug,
	}

	// Check if the user is authenticated so we can use it to
	// get the user's current roadmap progress.
	// If the user is not authenticated, we will return the roadmap
	// without the progress.
	authorization := r.Header.Get("Authorization")
	if authorization != "" {
		bearer := strings.Split(authorization, " ")
		if len(bearer) < 2 {
			a.handleError(w, r, cerrors.ErrUnauthorized)
			return
		}

		t := bearer[1]
		token, _ := a.application.AuthVerify(r.Context(), t)

		if token != nil {
			input.AccountID = token.AccountID
		}

		log := logger.FromContext(r.Context())
		log.Debug().Msgf("User ID: %d", token.AccountID)
	}

	output, err := a.application.GetRoadmapBySlug(r.Context(), *input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	a.render.OK(w, "Roadmap details.", output)
}
