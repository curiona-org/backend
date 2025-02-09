package backend

import (
	"context"
	"time"

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

	token, err := b.auth.Generate(result.id)
	if err != nil {
		return io.AuthOutput{}, err
	}

	output := io.AuthOutput{
		Created: result.created,
		Token:   token,
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
