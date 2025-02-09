package config

import (
	"time"
)

// Config is the global config for the application
type Config struct {
	AppName string
	AppEnv  string
	Port    string

	DBName                  string
	DBHost                  string
	DBPort                  int
	DBUser                  string
	DBPassword              string
	DBConnectionTimeout     int
	DBPoolMaxConnections    int
	DBPoolMinConnections    int
	DBPoolMaxConnLifetime   time.Duration
	DBPoolMaxConnIdleTime   time.Duration
	DBPoolHealthCheckPeriod time.Duration

	JWTSecretKey string
	JWTExpiresIn time.Duration

	LLMProvider string
	LLMAPIKey   string
	LLMModel    string

	GoogleClientID     string
	GoogleClientSecret string

	OTLPExporterEndpoint string
}

var config *Config

// Init initializes the config package
func Init() {
	config = &Config{
		AppName: LookupEnv("APP_NAME", "roadmap_backend"),
		AppEnv:  LookupEnv("APP_ENV", "local"),
		Port:    LookupEnv("PORT", "5000"),

		DBName:                  LookupEnv("DB_NAME", "roadmap"),
		DBHost:                  LookupEnv("DB_HOST", "localhost"),
		DBPort:                  LookupEnv("DB_PORT", 5432),
		DBUser:                  LookupEnv("DB_USER", "postgres"),
		DBPassword:              LookupEnv("DB_PASSWORD", ""),
		DBConnectionTimeout:     LookupEnv("DB_CONNECT_TIMEOUT", 5),
		DBPoolMinConnections:    LookupEnv("DB_POOL_MIN_CONNS", 0),
		DBPoolMaxConnections:    LookupEnv("DB_POOL_MAX_CONNS", 4),
		DBPoolMaxConnLifetime:   LookupEnv("DB_POOL_MAX_CONN_LIFETIME", time.Hour),
		DBPoolMaxConnIdleTime:   LookupEnv("DB_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute),
		DBPoolHealthCheckPeriod: LookupEnv("DB_POOL_HEALTH_CHECK_PERIOD", time.Minute),

		JWTSecretKey: LookupEnv("JWT_SECRET_KEY", "secret"),
		JWTExpiresIn: LookupEnv("JWT_EXPIRES_IN", time.Hour*24),

		LLMProvider: LookupEnv("LLM_PROVIDER", "deepseek"),

		LLMAPIKey:          LookupEnv("OPENAI_API_KEY", ""),
		LLMModel:           LookupEnv("OPENAI_MODEL", "gpt-4o-mini"),
		GoogleClientID:     LookupEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: LookupEnv("GOOGLE_CLIENT_SECRET", ""),

		OTLPExporterEndpoint: LookupEnv("OTLP_EXPORTER_ENDPOINT", "localhost:4317"),
	}
}

// GetConfig returns the global config
func GetConfig() *Config { return config }

func AppName() string     { return config.AppName }
func AppEnv() string      { return config.AppEnv }
func IsProduction() bool  { return config.AppEnv == "production" }
func IsDevelopment() bool { return config.AppEnv != "production" }
func Port() string        { return config.Port }

func DBName() string                         { return config.DBName }
func DBHost() string                         { return config.DBHost }
func DBPort() int                            { return config.DBPort }
func DBUser() string                         { return config.DBUser }
func DBPassword() string                     { return config.DBPassword }
func DBConnectionTimeout() int               { return config.DBConnectionTimeout }
func DBPoolMaxConnections() int              { return config.DBPoolMaxConnections }
func DBPoolMinConnections() int              { return config.DBPoolMinConnections }
func DBPoolMaxConnLifetime() time.Duration   { return config.DBPoolMaxConnLifetime }
func DBPoolMaxConnIdleTime() time.Duration   { return config.DBPoolMaxConnIdleTime }
func DBPoolHealthCheckPeriod() time.Duration { return config.DBPoolHealthCheckPeriod }

func JWTSecretKey() string        { return config.JWTSecretKey }
func JWTExpiresIn() time.Duration { return config.JWTExpiresIn }

func SetJWTSecretKey(key string)             { config.JWTSecretKey = key }
func SetJWTExpiresIn(duration time.Duration) { config.JWTExpiresIn = duration }

func LLMProvider() string { return config.LLMProvider }

func LLMAPIKey() string { return config.LLMAPIKey }
func LLMModel() string  { return config.LLMModel }

func GoogleClientID() string     { return config.GoogleClientID }
func GoogleClientSecret() string { return config.GoogleClientSecret }

func OTLPExporterEndpoint() string { return config.OTLPExporterEndpoint }
