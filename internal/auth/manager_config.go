package auth

import "time"

type ManagerConfig struct {
	AccessSecretKey  string
	AccessExpiresIn  time.Duration
	RefreshSecretKey string
	RefreshExpiresIn time.Duration
}
