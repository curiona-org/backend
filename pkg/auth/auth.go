package auth

import "time"

type Auth struct {
	Access  Token
	Refresh Token
}

type Config struct {
	AccessSecretKey  string
	AccessExpiresIn  time.Duration
	RefreshSecretKey string
	RefreshExpiresIn time.Duration

	// SymmetricKey for Paseto
	SymmetricKey string
}

type Strategy string

const (
	StrategyJWT    Strategy = "jwt"
	StrategyPaseto Strategy = "paseto"
)

func New(strategy Strategy, cfg *Config) *Auth {
	switch strategy {
	case StrategyJWT:
		return &Auth{
			Access:  NewJWT(cfg.AccessSecretKey, cfg.AccessExpiresIn),
			Refresh: NewJWT(cfg.RefreshSecretKey, cfg.RefreshExpiresIn),
		}
	case StrategyPaseto:
		// TODO: Implement Paseto
		return &Auth{
			Access:  NewNoop(),
			Refresh: NewNoop(),
		}
	default:
		return nil
	}
}
