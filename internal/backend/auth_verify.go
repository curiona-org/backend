package backend

import (
	"context"

	"github.com/roadmap-thesis/backend/pkg/auth"
)

func (b *backend) AuthVerify(ctx context.Context, token string) (*auth.Payload, error) {
	ctx, span := tracer.Start(ctx, "(*backend.AuthVerify)")
	defer span.End()

	return b.auth.Access.Parse(token)
}
