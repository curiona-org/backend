package application

import (
	"context"

	"github.com/roadmap-thesis/backend/pkg/auth"
)

//nolint:revive
func (app *application) AuthVerify(ctx context.Context, token string) (*auth.Payload, error) {
	return app.auth.Access.Parse(token)
}
