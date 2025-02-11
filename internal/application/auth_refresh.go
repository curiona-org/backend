package application

import (
	"context"
	"errors"
	"time"

	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/internal/io"
)

// AuthRefresh refreshes the access token using the refresh token and returns a new access token and a rotated refresh token.
// The old session will be blocked to prevent replay attacks while a new session is created.
// Currently, it supports multiple sessions per user (e.g., multiple devices).
func (app *application) AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error) {
	ctx, span := tracer.Start(ctx, "(*application.AuthRefresh)")
	defer span.End()

	payload, err := app.auth.Refresh.Parse(input.Token)
	if err != nil {
		return io.AuthRefreshOutput{}, err
	}

	var accessToken, refreshToken string
	var refreshExpiresAt time.Time
	err = app.repository.Session.UpdateByRefreshToken(ctx, input.Token, func(traceCtx context.Context, session *domain.Session) (bool, error) {
		if session.Blocked {
			err := app.repository.Session.Delete(traceCtx, session.ID)
			if err != nil {
				return false, err
			}
			return false, errors.Join(apperrors.Unauthorized(), errors.New("session is blocked"))
		}

		if time.Now().After(session.ExpiresAt) {
			return false, errors.Join(apperrors.Unauthorized(), errors.New("refresh token expired"))
		}

		accessToken, err = app.auth.Access.Generate(payload.ID)
		if err != nil {
			return false, err
		}

		// Block the old session to prevent replay attacks
		session.MarkAsBlocked()

		// Rotate the refresh token
		refreshToken, err = app.auth.Refresh.Generate(payload.ID)
		if err != nil {
			return false, err
		}

		newSession := domain.NewSession(
			payload.ID,
			refreshToken,
			session.UserAgent,
			session.ClientIP,
			app.auth.Refresh.ExpiresAt(),
		)

		_, err = app.repository.Session.Save(traceCtx, newSession)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return io.AuthRefreshOutput{}, errors.Join(apperrors.Unauthorized(), errors.New("session not found"))
		}
		return io.AuthRefreshOutput{}, err
	}

	return io.AuthRefreshOutput{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  time.Now().Add(app.auth.Access.ExpiresIn()),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}
