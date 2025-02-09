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
	setEnv("DB_NAME", "roadmap")
	setEnv("DB_HOST", "localhost")
	setEnv("DB_PORT", "5432")
	setEnv("DB_USER", "postgres")
	setEnv("DB_PASSWORD", "")
	setEnv("DB_CONNECTION_TIMEOUT", "5")
	setEnv("DB_POOL_MIN_CONNECTIONS", "0")
	setEnv("DB_POOL_MAX_CONNECTIONS", "4")
	setEnv("DB_POOL_MAX_CONN_LIFETIME", "1h")
	setEnv("DB_POOL_MAX_CONN_IDLETIME", "30m")
	setEnv("DB_POOL_HEALTH_CHECK_PERIOD", "1m")
	setEnv("JWT_SECRET_KEY", "test_secret")
	setEnv("JWT_EXPIRES_IN", "48h")
	setEnv("OPENAI_API_KEY", "test_api_key")
	setEnv("OPENAI_MODEL", "gpt-4")

	defer unsetEnv("APP_NAME")
	defer unsetEnv("APP_ENV")
	defer unsetEnv("PORT")
	defer unsetEnv("DB_NAME")
	defer unsetEnv("DB_HOST")
	defer unsetEnv("DB_PORT")
	defer unsetEnv("DB_USER")
	defer unsetEnv("DB_PASSWORD")
	defer unsetEnv("DB_CONNECTION_TIMEOUT")
	defer unsetEnv("DB_POOL_MIN_CONNECTIONS")
	defer unsetEnv("DB_POOL_MAX_CONNECTIONS")
	defer unsetEnv("DB_POOL_MAX_CONN_LIFETIME")
	defer unsetEnv("DB_POOL_MAX_CONN_IDLETIME")
	defer unsetEnv("DB_POOL_HEALTH_CHECK_PERIOD")
	defer unsetEnv("JWT_SECRET_KEY")
	defer unsetEnv("JWT_EXPIRES_IN")
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
		{"DBName", "roadmap", cfg.DBName},
		{"DBNameFunc", "roadmap", config.DBName()},
		{"DBHost", "localhost", cfg.DBHost},
		{"DBHostFunc", "localhost", config.DBHost()},
		{"DBPort", 5432, cfg.DBPort},
		{"DBPortFunc", 5432, config.DBPort()},
		{"DBUser", "postgres", cfg.DBUser},
		{"DBUserFunc", "postgres", config.DBUser()},
		{"DBPassword", "", cfg.DBPassword},
		{"DBPasswordFunc", "", config.DBPassword()},
		{"DBConnectionTimeout", 5, cfg.DBConnectionTimeout},
		{"DBConnectionTimeoutFunc", 5, config.DBConnectionTimeout()},
		{"DBPoolMinConnections", 0, cfg.DBPoolMinConnections},
		{"DBPoolMinConnectionsFunc", 0, config.DBPoolMinConnections()},
		{"DBPoolMaxConnections", 4, cfg.DBPoolMaxConnections},
		{"DBPoolMaxConnectionsFunc", 4, config.DBPoolMaxConnections()},
		{"DBPoolMaxConnLifetime", time.Hour, cfg.DBPoolMaxConnLifetime},
		{"DBPoolMaxConnLifetimeFunc", time.Hour, config.DBPoolMaxConnLifetime()},
		{"DBPoolMaxConnIdleTime", 30 * time.Minute, cfg.DBPoolMaxConnIdleTime},
		{"DBPoolMaxConnIdleTimeFunc", 30 * time.Minute, config.DBPoolMaxConnIdleTime()},
		{"DBPoolHealthCheckPeriod", time.Minute, cfg.DBPoolHealthCheckPeriod},
		{"DBPoolHealthCheckPeriodFunc", time.Minute, config.DBPoolHealthCheckPeriod()},
		{"JWTSecretKey", "test_secret", cfg.JWTSecretKey},
		{"JWTSecretKeyFunc", "test_secret", config.JWTSecretKey()},
		{"JWTExpiresIn", 48 * time.Hour, cfg.JWTExpiresIn},
		{"JWTExpiresInFunc", 48 * time.Hour, config.JWTExpiresIn()},
		{"LLMAPIKey", "test_api_key", cfg.LLMAPIKey},
		{"LLMAPIKeyFunc", "test_api_key", config.LLMAPIKey()},
		{"LLMModel", "gpt-4", cfg.LLMModel},
		{"LLMModelFunc", "gpt-4", config.LLMModel()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.actual)
		})
	}
}
