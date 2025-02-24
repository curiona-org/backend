package auth

import (
	"time"
)

type Auth struct {
	cfg *Config
}

type Config struct {
	AccessSecretKey  string
	AccessExpiresIn  time.Duration
	RefreshSecretKey string
	RefreshExpiresIn time.Duration
}

func New(cfg *Config) *Auth {
	return &Auth{
		cfg: cfg,
	}
}

func (a *Auth) NewAccessToken(id int) Token {
	return NewToken(a.cfg.AccessSecretKey, id, a.cfg.AccessExpiresIn)
}

func (a *Auth) NewRefreshToken(id int) Token {
	return NewToken(a.cfg.RefreshSecretKey, id, a.cfg.RefreshExpiresIn)
}

func (a *Auth) VerifyAccessToken(tokenStr string) (Token, error) {
	return TokenUnmarshal(a.cfg.AccessSecretKey, tokenStr)
}

func (a *Auth) VerifyRefreshToken(tokenStr string) (Token, error) {
	return TokenUnmarshal(a.cfg.RefreshSecretKey, tokenStr)
}
