package app

import (
	"context"
	"errors"
	"time"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/pkg/cerrors"
)

// AuthRefresh refreshes the access token using the refresh token and returns a new access token and a rotated refresh token.
// The old session will be blocked to prevent replay attacks while a new session is created.
// Currently, it supports multiple sessions per user (e.g., multiple devices).
func (app *application) AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.AuthRefresh)")
	defer span.End()

	payload, err := app.auth.Refresh.Parse(input.Token)
	if err != nil {
		return io.AuthRefreshOutput{}, err
	}

	var accessToken, refreshToken string
	err = app.repository.Session.Renew(ctx, input.Token, func(session *domain.Session) (bool, error) {
		if session.Blocked {
			return false, cerrors.Wrap(cerrors.Unauthorized(), domain.ErrSessionIsBlocked)
		}

		if time.Now().After(session.ExpiresAt) {
			return false, cerrors.Wrap(cerrors.Unauthorized(), domain.ErrSessionExpired)
		}

		accessToken, err = app.auth.Access.Generate(payload.ID)
		if err != nil {
			return false, err
		}

		// Rotate the refresh token
		refreshToken, err = app.auth.Refresh.Generate(payload.ID)
		if err != nil {
			return false, err
		}

		session.Renew(
			refreshToken,
			input.UserAgent,
			input.ClientIP,
			app.auth.Refresh.ExpiresAt(),
		)
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return io.AuthRefreshOutput{}, cerrors.Wrap(cerrors.Unauthorized(), err)
		}
		return io.AuthRefreshOutput{}, err
	}

	return io.AuthRefreshOutput{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  app.auth.Access.ExpiresAt(),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresIn: int(app.auth.Refresh.ExpiresIn().Seconds()),
		RefreshTokenExpiresAt: app.auth.Refresh.ExpiresAt(),
	}, nil
}
