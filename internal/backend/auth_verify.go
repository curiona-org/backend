package backend

import (
	"context"

	"github.com/roadmap-thesis/backend/pkg/auth"
)

func (b *backend) AuthVerify(ctx context.Context, token string) (*auth.Payload, error) {
	return b.auth.Access.Parse(token)
}
