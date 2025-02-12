package application

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/io"
	"github.com/roadmap-thesis/backend/pkg/str"
	"go.opentelemetry.io/otel/codes"
)

func (app *application) authGoogle(ctx context.Context, input io.AuthInput) (registrationResult, error) {
	ctx, span := tracer.Start(ctx, "(*application.authGoogle)")
	defer span.End()

	user, err := app.googleOAuth.Verify(ctx, input.OAuthToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return registrationResult{}, apperrors.Unauthorized()
	}

	return app.authEmailPassword(ctx, io.AuthInput{
		Name:                user.Name,
		Email:               user.Email,
		Password:            str.Random(32),
		Avatar:              user.Avatar,
		IgnorePasswordCheck: true,
	})
}
