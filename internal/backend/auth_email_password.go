package backend

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/internal/io"
	"github.com/roadmap-thesis/backend/pkg/apperrors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (b *backend) authEmailPassword(ctx context.Context, input io.AuthInput) (registrationResult, error) {
	ctx, span := tracer.Start(ctx, "(*backend.authEmailPassword)")
	defer span.End()

	existingAccount, err := b.repository.Account.GetByEmail(ctx, input.Email)
	if err != nil && err != domain.ErrAccountNotFound {
		return registrationResult{}, err
	}

	// sign in if account already exists
	if !existingAccount.IsZero() {
		span.SetAttributes(attribute.Bool("create_account", false))

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

		matched := existingAccount.CheckPassword(input.Password)
		if !matched {
			return registrationResult{}, apperrors.InvalidCredentials()
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
	profile := domain.NewProfile(input.Name, input.Avatar)
	account, err := domain.NewAccount(input.Email, input.Password, profile)
	if err != nil {
		return registrationResult{}, err
	}

	createdAccount, err := b.repository.Account.Save(ctx, account)
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
