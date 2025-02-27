package admin

import "context"

func (app *adminApplication) IsAdmin(ctx context.Context, accountID int) (bool, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.IsAdmin)")
	defer span.End()

	account, err := app.repository.Account.GetByID(ctx, accountID)
	if err != nil {
		return false, err
	}

	return account.IsAdmin, nil
}
