package app

import (
	"context"

	"github.com/curiona-org/backend/internal/auth"
)

func (app *application) AuthVerify(ctx context.Context, token string) (*auth.Token, error) {
	return app.auth.VerifyAccessToken(token)
}
