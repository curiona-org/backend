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
	setEnv("ACCESS_SECRET_KEY", "test_secret")
	setEnv("ACCESS_EXPIRES_IN", "5m")
	setEnv("REFRESH_SECRET_KEY", "test_secret")
	setEnv("REFRESH_EXPIRES_IN", "720h")
	setEnv("LLM_API_KEY", "test_api_key")
	setEnv("LLM_MODEL", "gpt-4")
	setEnv("GOOGLE_CLIENT_ID", "test_google_id")
	setEnv("GOOGLE_CLIENT_SECRET", "test_google_secret")
	setEnv("OTLP_EXPORTER_ENDPOINT", "http://localhost:4317")

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
	defer unsetEnv("ACCESS_SECRET_KEY")
	defer unsetEnv("ACCESS_EXPIRES_IN")
	defer unsetEnv("REFRESH_SECRET_KEY")
	defer unsetEnv("REFRESH_EXPIRES_IN")
	defer unsetEnv("LLM_API_KEY")
	defer unsetEnv("LLM_MODEL")
	defer unsetEnv("GOOGLE_CLIENT_ID")
	defer unsetEnv("GOOGLE_CLIENT_SECRET")
	defer unsetEnv("OTLP_EXPORTER_ENDPOINT")

	config.Init()

	testCases := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"AppName", "test_app", config.AppName()},
		{"AppEnv", "test", config.AppEnv()},
		{"Port", "8080", config.Port()},
		{"DBName", "roadmap", config.DBName()},
		{"DBHost", "localhost", config.DBHost()},
		{"DBPort", 5432, config.DBPort()},
		{"DBUser", "postgres", config.DBUser()},
		{"DBPassword", "", config.DBPassword()},
		{"DBConnectionTimeout", 5, config.DBConnectionTimeout()},
		{"DBPoolMinConnections", 0, config.DBPoolMinConnections()},
		{"DBPoolMaxConnections", 4, config.DBPoolMaxConnections()},
		{"DBPoolMaxConnLifetime", time.Hour, config.DBPoolMaxConnLifetime()},
		{"DBPoolMaxConnIdleTime", 30 * time.Minute, config.DBPoolMaxConnIdleTime()},
		{"DBPoolHealthCheckPeriod", time.Minute, config.DBPoolHealthCheckPeriod()},
		{"AccessSecretKey", "test_secret", config.AccessSecretKey()},
		{"AccessExpiresIn", 5 * time.Minute, config.AccessExpiresIn()},
		{"RefreshSecretKey", "test_secret", config.RefreshSecretKey()},
		{"RefreshExpiresIn", 24 * time.Hour * 30, config.RefreshExpiresIn()},
		{"LLMAPIKey", "test_api_key", config.LLMAPIKey()},
		{"LLMModel", "gpt-4", config.LLMModel()},
		{"GoogleClientID", "test_google_id", config.GoogleClientID()},
		{"GoogleClientSecret", "test_google_secret", config.GoogleClientSecret()},
		{"OTLPExporterEndpoint", "http://localhost:4317", config.OTLPExporterEndpoint()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.actual)
		})
	}
}
