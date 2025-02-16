package application

import (
	"context"
	"time"

	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/internal/io"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.Auth)", trace.WithAttributes(attribute.String("email", input.Email)))
	defer span.End()

	var result registrationResult
	var err error
	if input.OAuthToken != "" {
		result, err = app.authGoogle(ctx, input)
	} else {
		input.Provider = domain.AccountProviderEmail
		result, err = app.authEmailPassword(ctx, input)
	}

	if err != nil {
		return io.AuthOutput{}, err
	}

	accessToken, err := app.auth.Access.Generate(result.id)
	if err != nil {
		return io.AuthOutput{}, err
	}

	refreshToken, err := app.auth.Refresh.Generate(result.id)
	if err != nil {
		return io.AuthOutput{}, err
	}

	refreshExpiresAt := app.auth.Refresh.ExpiresAt()

	newSession := domain.NewSession(
		result.id,
		refreshToken,
		input.UserAgent,
		input.ClientIP,
		refreshExpiresAt,
	)

	_, err = app.repository.Session().Save(ctx, newSession)
	if err != nil {
		return io.AuthOutput{}, err
	}

	output := io.AuthOutput{
		Created:               result.created,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  app.auth.Access.ExpiresAt(),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresIn: int(app.auth.Refresh.ExpiresIn().Seconds()),
		RefreshTokenExpiresAt: app.auth.Refresh.ExpiresAt(),
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
