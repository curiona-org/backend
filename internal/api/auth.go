package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) Auth(w http.ResponseWriter, r *http.Request) {
	var input io.AuthInput

	if err := a.Bind(r.Body, &input); err != nil {
		a.handleError(w, r, cerrors.ErrInvalidData)
		return
	}

	if err := a.validator.Validate(&input); err != nil {
		a.handleError(w, r, err)
		return
	}

	input.UserAgent = r.UserAgent()
	input.ClientIP = r.RemoteAddr
	output, err := a.application.Auth(r.Context(), input)
	if err != nil {
		a.handleError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/",
		MaxAge:   output.RefreshTokenExpiresIn,
		Expires:  output.RefreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	if output.Created {
		a.render.Created(w, "Successfully registered.", output)
		return
	}

	a.render.OK(w, "Successfully logged in.", output)
}
