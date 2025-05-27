package api

import (
	"net/http"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/config"
	"github.com/curiona-org/backend/internal/logger"
)

func (a *API) AuthRefresh(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get refresh token from cookie")
		a.handleError(w, r, cerrors.ErrUnauthorized)
		return
	}

	log.Debug().Str("refresh_token", refreshToken.Value).Msg("Received refresh token")

	output, err := a.application.AuthRefresh(r.Context(), io.AuthRefreshInput{
		Token:     refreshToken.Value,
		UserAgent: r.UserAgent(),
		ClientIP:  r.RemoteAddr,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to refresh token")
		a.handleError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Domain:   config.FrontendDomain(),
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/",
		MaxAge:   output.RefreshTokenExpiresIn,
		Expires:  output.RefreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   config.IsProduction(),
		SameSite: http.SameSiteNoneMode,
	})

	a.render.OK(w, "Successfully refreshed token.", output)
}
