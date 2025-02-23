package app

import (
	"context"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain/object"
	"github.com/curiona-org/backend/pkg/cerrors"
	"github.com/curiona-org/backend/pkg/str"
	"go.opentelemetry.io/otel/codes"
)

func (app *application) authGoogle(ctx context.Context, input io.AuthInput) (registrationResult, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.authGoogle)")
	defer span.End()

	user, err := app.googleOAuth.Verify(ctx, input.OAuthToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return registrationResult{}, cerrors.Wrap(cerrors.Unauthorized(), err)
	}

	return app.authEmailPassword(ctx, io.AuthInput{
		Name:     user.Name,
		Email:    user.Email,
		Password: str.Random(32),
		Avatar:   user.Avatar,

		Provider:            object.AccountProviderGoogle,
		IgnorePasswordCheck: true,
	})
}
