package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/stretchr/testify/assert"
)

func setEnv(key, value string) {
	os.Setenv(key, value)
}

func unsetEnv(key string) {
	os.Unsetenv(key)
}

func TestConfig_Init(t *testing.T) {
	t.Parallel()
	setEnv("APP_NAME", "test_app")
	setEnv("APP_ENV", "test")
	setEnv("PORT", "8080")
	setEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname")
	setEnv("DATABASE_MAX_CONNS", "10")
	setEnv("DATABASE_MIN_CONNS", "1")
	setEnv("DATABASE_MAX_CONN_LIFETIME", "2h")
	setEnv("DATABASE_MAX_CONN_IDLETIME", "1h")
	setEnv("DATABASE_HEALTH_CHECK_PERIOD", "30m")
	setEnv("DATABASE_DEFAULT_CONNECTION_TIMEOUT", "10s")
	setEnv("JWT_SECRET_KEY", "test_secret")
	setEnv("JWT_SECRET_EXPIRES_IN", "48h")
	setEnv("OPENAI_API_KEY", "test_api_key")
	setEnv("OPENAI_MODEL", "gpt-4")

	defer unsetEnv("APP_NAME")
	defer unsetEnv("APP_ENV")
	defer unsetEnv("PORT")
	defer unsetEnv("DATABASE_URL")
	defer unsetEnv("DATABASE_MAX_CONNS")
	defer unsetEnv("DATABASE_MIN_CONNS")
	defer unsetEnv("DATABASE_MAX_CONN_LIFETIME")
	defer unsetEnv("DATABASE_MAX_CONN_IDLETIME")
	defer unsetEnv("DATABASE_HEALTH_CHECK_PERIOD")
	defer unsetEnv("DATABASE_DEFAULT_CONNECTION_TIMEOUT")
	defer unsetEnv("JWT_SECRET_KEY")
	defer unsetEnv("JWT_SECRET_EXPIRES_IN")
	defer unsetEnv("OPENAI_API_KEY")
	defer unsetEnv("OPENAI_MODEL")

	config.Init()

	cfg := config.GetConfig()
	testCases := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"AppName", "test_app", cfg.AppName},
		{"AppNameFunc", "test_app", config.AppName()},
		{"AppEnv", "test", cfg.AppEnv},
		{"AppEnvFunc", "test", config.AppEnv()},
		{"Port", "8080", cfg.Port},
		{"PortFunc", "8080", config.Port()},
		{"DatabaseURL", "postgres://user:pass@localhost:5432/dbname", cfg.DatabaseURL},
		{"DatabaseURLFunc", "postgres://user:pass@localhost:5432/dbname", config.DatabaseURL()},
		{"DatabaseMaxConns", int32(10), cfg.DatabaseMaxConns},
		{"DatabaseMaxConnsFunc", int32(10), config.DatabaseMaxConns()},
		{"DatabaseMinConns", int32(1), cfg.DatabaseMinConns},
		{"DatabaseMinConnsFunc", int32(1), config.DatabaseMinConns()},
		{"DatabaseMaxConnLifetime", 2 * time.Hour, cfg.DatabaseMaxConnLifetime},
		{"DatabaseMaxConnLifetimeFunc", 2 * time.Hour, config.DatabaseMaxConnLifetime()},
		{"DatabaseMaxConnIdleTime", 1 * time.Hour, cfg.DatabaseMaxConnIdleTime},
		{"DatabaseMaxConnIdleTimeFunc", 1 * time.Hour, config.DatabaseMaxConnIdleTime()},
		{"DatabaseHealthCheckPeriod", 30 * time.Minute, cfg.DatabaseHealthCheckPeriod},
		{"DatabaseHealthCheckPeriodFunc", 30 * time.Minute, config.DatabaseHealthCheckPeriod()},
		{"DatabaseDefaultConnectionTimeout", 10 * time.Second, cfg.DatabaseDefaultConnectionTimeout},
		{"DatabaseDefaultConnectionTimeoutFunc", 10 * time.Second, config.DatabaseDefaultConnectionTimeout()},
		{"JWTSecretKey", "test_secret", cfg.JWTSecretKey},
		{"JWTSecretKeyFunc", "test_secret", config.JWTSecretKey()},
		{"JWTSecretExpiresIn", 48 * time.Hour, cfg.JWTSecretExpiresIn},
		{"JWTSecretExpiresInFunc", 48 * time.Hour, config.JWTSecretExpiresIn()},
		{"OpenAiAPIKey", "test_api_key", cfg.OpenAiAPIKey},
		{"OpenAiAPIKeyFunc", "test_api_key", config.OpenAiAPIKey()},
		{"OpenAiModel", "gpt-4", cfg.OpenAiModel},
		{"OpenAiModelFunc", "gpt-4", config.OpenAiModel()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.actual)
		})
	}
}
