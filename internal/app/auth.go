package app

import (
	"context"
	"time"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.Auth)", trace.WithAttributes(attribute.String("email", input.Email)))
	defer span.End()

	input.Email = auth.NormalizeEmail(input.Email)

	var result registrationResult
	var err error
	if input.OAuthToken != "" {
		result, err = app.authGoogle(ctx, input)
	} else {
		input.Method = auth.MethodEmail
		result, err = app.authEmailPassword(ctx, input)
	}

	if err != nil {
		return io.AuthOutput{}, err
	}

	accessToken := app.auth.NewAccessToken(result.id)
	refreshToken := app.auth.NewRefreshToken(result.id)

	accessTokenStr, err := accessToken.Marshal()
	if err != nil {
		return io.AuthOutput{}, err
	}

	refreshTokenStr, err := refreshToken.Marshal()
	if err != nil {
		return io.AuthOutput{}, err
	}

	newSession := domain.NewSession(
		result.id,
		refreshTokenStr,
		input.UserAgent,
		input.ClientIP,
		refreshToken.ExpiresAt,
	)

	_, err = app.repository.Session.Save(ctx, newSession)
	if err != nil {
		return io.AuthOutput{}, err
	}

	output := io.AuthOutput{
		Created:               result.created,
		AccessToken:           accessTokenStr,
		AccessTokenExpiresAt:  accessToken.ExpiresAt,
		RefreshToken:          refreshTokenStr,
		RefreshTokenExpiresIn: int(refreshToken.ExpiresIn().Seconds()),
		RefreshTokenExpiresAt: refreshToken.ExpiresAt,
		Account: io.AuthOutputAccount{
			ID:       result.id,
			Email:    result.email,
			Name:     result.name,
			Avatar:   result.avatar,
			JoinedAt: result.joinedAt,
		},
	}

	return output, nil
}

type registrationResult struct {
	id       int
	created  bool
	name     string
	avatar   string
	email    string
	joinedAt time.Time
}
