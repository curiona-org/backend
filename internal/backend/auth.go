package backend

import (
	"context"
	"time"

	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/internal/io"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (b *backend) Auth(ctx context.Context, input io.AuthInput) (io.AuthOutput, error) {
	ctx, span := tracer.Start(ctx, "(*backend.Auth)", trace.WithAttributes(attribute.String("email", input.Email)))
	defer span.End()

	var result registrationResult
	var err error
	if input.OAuthToken != "" {
		result, err = b.authGoogle(ctx, input)
	} else {
		result, err = b.authEmailPassword(ctx, input)
	}

	if err != nil {
		return io.AuthOutput{}, err
	}

	accessToken, err := b.auth.Access.Generate(result.id)
	if err != nil {
		return io.AuthOutput{}, err
	}

	refreshToken, err := b.auth.Refresh.Generate(result.id)
	if err != nil {
		return io.AuthOutput{}, err
	}

	refreshExpiresAt := b.auth.Refresh.ExpiresAt()

	newSession := domain.NewSession(
		result.id,
		refreshToken,
		input.UserAgent,
		input.ClientIP,
		refreshExpiresAt,
	)

	_, err = b.repository.Session.Save(ctx, newSession)
	if err != nil {
		return io.AuthOutput{}, err
	}

	output := io.AuthOutput{
		Created:               result.created,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  b.auth.Access.ExpiresAt(),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
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
