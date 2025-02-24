package auth

type contextKey string

const (
	// ContextKey is the key used to store the authpayload in a context.
	ContextKey contextKey = "auth_payload"
)

func (k contextKey) String() string {
	return string(k)
}
