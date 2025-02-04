package backend

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/io"
	"github.com/roadmap-thesis/backend/pkg/apperrors"
	"github.com/roadmap-thesis/backend/pkg/str"
	"go.opentelemetry.io/otel/codes"
)

func (b *backend) authGoogle(ctx context.Context, input io.AuthInput) (registrationResult, error) {
	ctx, span := tracer.Start(ctx, "(*backend.authGoogle)")
	defer span.End()

	user, err := b.googleOAuth.Verify(ctx, input.OAuthToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return registrationResult{}, apperrors.Unauthorized()
	}

	return b.authEmailPassword(ctx, io.AuthInput{
		Name:                user.Name,
		Email:               user.Email,
		Password:            str.Random(32),
		Avatar:              user.Avatar,
		IgnorePasswordCheck: true,
	})
}
