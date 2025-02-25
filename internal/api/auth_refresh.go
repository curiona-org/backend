package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
)

func (a *API) AuthRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		a.handleError(w, r, cerrors.ErrUnauthorized)
		return
	}

	output, err := a.application.AuthRefresh(r.Context(), io.AuthRefreshInput{
		Token:     refreshToken.Value,
		UserAgent: r.UserAgent(),
		ClientIP:  r.RemoteAddr,
	})
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

	a.render.OK(w, "Successfully refreshed token.", output)
}
