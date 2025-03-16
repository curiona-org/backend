package app

import (
	"context"
)

func (app *application) HealthCheck(ctx context.Context) bool {
	return app.repository.Ping(ctx)
}
