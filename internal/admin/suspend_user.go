package admin

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *adminApplication) SuspendUser(ctx context.Context, accountID int) error {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.SuspendUser)")
	defer span.End()

	err := app.repository.Account.Update(ctx, accountID, func(account *domain.Account) (bool, error) {
		if account.IsSuspended {
			return false, nil
		}

		account.Suspend()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return cerrors.ErrNotFound.Msg("account")
		}
		return err
	}

	return nil
}
