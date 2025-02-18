package object

// AccountProvider represents the type of account provider.
type AccountProvider string

const (
	// AccountProviderEmail uses email & password.
	AccountProviderEmail AccountProvider = "email"
	// AccountProviderGoogle uses Google OAuth.
	AccountProviderGoogle AccountProvider = "google"
)
