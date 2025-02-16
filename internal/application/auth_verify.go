package application

import (
	"context"

	"github.com/roadmap-thesis/backend/pkg/auth"
)

func (app *application) AuthVerify(ctx context.Context, token string) (*auth.Payload, error) { //nolint:revive
	return app.auth.Access.Parse(token)
}
