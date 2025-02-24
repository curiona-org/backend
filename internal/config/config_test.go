package config_test

import (
	"testing"
	"time"

	"github.com/curiona-org/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Init(t *testing.T) {
	t.Setenv("APP_NAME", "test_app")
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "8080")
	t.Setenv("DB_NAME", "roadmap")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_CONNECTION_TIMEOUT", "5")
	t.Setenv("DB_POOL_MIN_CONNECTIONS", "0")
	t.Setenv("DB_POOL_MAX_CONNECTIONS", "4")
	t.Setenv("DB_POOL_MAX_CONN_LIFETIME", "1h")
	t.Setenv("DB_POOL_MAX_CONN_IDLETIME", "30m")
	t.Setenv("DB_POOL_HEALTH_CHECK_PERIOD", "1m")
	t.Setenv("REDIS_DB", "0")
	t.Setenv("REDIS_NETWORK", "tcp")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("REDIS_USERNAME", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("ACCESS_SECRET_KEY", "test_secret")
	t.Setenv("ACCESS_EXPIRES_IN", "5m")
	t.Setenv("REFRESH_SECRET_KEY", "test_secret")
	t.Setenv("REFRESH_EXPIRES_IN", "720h")
	t.Setenv("LLM_API_KEY", "test_api_key")
	t.Setenv("LLM_MODEL", "gpt-4")
	t.Setenv("GOOGLE_CLIENT_ID", "test_google_id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test_google_secret")
	t.Setenv("OTLP_EXPORTER_ENDPOINT", "http://localhost:4317")

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
		{"RedisDB", 0, config.RedisDB()},
		{"RedisNetwork", "tcp", config.RedisNetwork()},
		{"RedisAddr", "localhost:6379", config.RedisAddr()},
		{"RedisUsername", "", config.RedisUsername()},
		{"RedisPassword", "", config.RedisPassword()},
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
