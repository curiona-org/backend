package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (app *application) authEmailPassword(ctx context.Context, input io.AuthInput) (registrationResult, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.authEmailPassword)")
	defer span.End()

	existingAccount, err := app.repository.Account.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrAccountNotFound) {
		return registrationResult{}, err
	}

	if existingAccount.IsSuspended {
		return registrationResult{}, cerrors.ErrAccountSuspended
	}

	// sign in if account already exists
	if !existingAccount.IsZero() {
		span.SetAttributes(attribute.Bool("create_account", false))

		// check if user already registered with a different method
		if existingAccount.Method != input.Method {
			return registrationResult{}, cerrors.ErrSignUpDifferentMethod
		}

		// ignore password check if user is signing in with google
		if input.IgnorePasswordCheck {
			return registrationResult{
				id:       existingAccount.ID,
				created:  false,
				email:    existingAccount.Email,
				name:     existingAccount.Profile.Name,
				avatar:   existingAccount.Profile.Avatar,
				joinedAt: existingAccount.CreatedAt,
			}, nil
		}

		plainPassword := auth.NewPassword(input.Password)
		matched := existingAccount.CheckPassword(plainPassword)
		if !matched {
			return registrationResult{}, cerrors.ErrInvalidCredentials
		}

		return registrationResult{
			id:       existingAccount.ID,
			created:  false,
			email:    existingAccount.Email,
			name:     existingAccount.Profile.Name,
			avatar:   existingAccount.Profile.Avatar,
			joinedAt: existingAccount.CreatedAt,
		}, nil
	}

	span.SetAttributes(attribute.Bool("create_account", true))

	password := auth.NewPassword(input.Password)

	if err := password.Validate(); err != nil {
		return registrationResult{}, err
	}

	profile := domain.NewProfile(input.Name, input.Avatar)
	account, err := domain.NewAccount(input.Email, password, input.Method, profile)
	if err != nil {
		return registrationResult{}, err
	}

	// Hash the password before saving
	if err := account.HashPassword(); err != nil {
		return registrationResult{}, err
	}

	createdAccount, err := app.repository.Account.Save(ctx, account)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return registrationResult{}, err
	}

	return registrationResult{
		id:       createdAccount.ID,
		created:  true,
		email:    createdAccount.Email,
		name:     createdAccount.Profile.Name,
		avatar:   createdAccount.Profile.Avatar,
		joinedAt: createdAccount.CreatedAt,
	}, nil
}
