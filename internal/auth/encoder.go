package auth

import "github.com/curiona-org/backend/internal/auth/jwt"

// TokenEncoder is an interface that defines methods for encoding and decoding
// tokens with a map of claims. Implementations of this interface should provide
// mechanisms to marshal claims into a token string and unmarshal a token string
// back into a map of claims.
type TokenEncoder interface {
	// Marshal encodes the given claims into a token string.
	// Returns the encoded token string or an error if encoding fails.
	Marshal(claims map[string]any) (string, error)

	// Unmarshal decodes the given token string into the provided output map.
	// Returns an error if decoding fails.
	Unmarshal(token string, out map[string]any) error
}

var _ TokenEncoder = (*jwt.JWT)(nil)
