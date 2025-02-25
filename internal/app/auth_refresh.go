package app

import (
	"context"
	"errors"
	"time"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

// AuthRefresh refreshes the access token using the refresh token and returns a new access token and a rotated refresh token.
// The old session will be blocked to prevent replay attacks while a new session is created.
// Currently, it supports multiple sessions per user (e.g., multiple devices).
func (app *application) AuthRefresh(ctx context.Context, input io.AuthRefreshInput) (io.AuthRefreshOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.AuthRefresh)")
	defer span.End()

	token, err := app.auth.VerifyRefreshToken(input.Token)
	if err != nil {
		return io.AuthRefreshOutput{}, err
	}

	accessToken := app.auth.NewAccessToken(token.AccountID)
	accessTokenStr, err := accessToken.Marshal()
	if err != nil {
		return io.AuthRefreshOutput{}, err
	}

	refreshToken := app.auth.NewRefreshToken(token.AccountID)
	var refreshTokenStr string
	err = app.repository.Session.Renew(ctx, input.Token, func(session *domain.Session) (bool, error) {
		if session.Blocked {
			return false, cerrors.Wrap(cerrors.Unauthorized, domain.ErrSessionIsBlocked)
		}

		if time.Now().After(session.ExpiresAt) {
			return false, cerrors.Wrap(cerrors.Unauthorized, domain.ErrSessionExpired)
		}

		// Rotate the refresh token
		refreshTokenStr, err = refreshToken.Marshal()
		if err != nil {
			return false, err
		}

		session.Renew(
			refreshTokenStr,
			input.UserAgent,
			input.ClientIP,
			refreshToken.ExpiresAt,
		)
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return io.AuthRefreshOutput{}, cerrors.Wrap(cerrors.Unauthorized, err)
		}
		return io.AuthRefreshOutput{}, err
	}

	return io.AuthRefreshOutput{
		AccessToken:           accessTokenStr,
		AccessTokenExpiresAt:  accessToken.ExpiresAt,
		RefreshToken:          refreshTokenStr,
		RefreshTokenExpiresIn: int(refreshToken.ExpiresIn().Seconds()),
		RefreshTokenExpiresAt: refreshToken.ExpiresAt,
	}, nil
}
