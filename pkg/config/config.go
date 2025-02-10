package config

import (
	"time"

	"github.com/roadmap-thesis/backend/pkg/llm"
)

// Config is the global config for the application
type Config struct {
	appName string
	appEnv  string
	port    string

	dbName                  string
	dbHost                  string
	dbPort                  int
	dbUser                  string
	dbPassword              string
	dbConnectionTimeout     int
	dbPoolMaxConnections    int
	dbPoolMinConnections    int
	dbPoolMaxConnLifetime   time.Duration
	dbPoolMaxConnIdleTime   time.Duration
	dbPoolHealthCheckPeriod time.Duration

	jwtSecretKey string
	jwtExpiresIn time.Duration

	llmProvider string
	llmAPIKey   string
	llmModel    string

	googleClientID     string
	googleClientSecret string

	otlpExporterEndpoint string
}

var config *Config

// Init initializes the config package
func Init() {
	config = &Config{
		appName: LookupEnv("APP_NAME", "roadmap_backend"),
		appEnv:  LookupEnv("APP_ENV", "local"),
		port:    LookupEnv("PORT", "5000"),

		dbName:                  LookupEnv("DB_NAME", "roadmap"),
		dbHost:                  LookupEnv("DB_HOST", "localhost"),
		dbPort:                  LookupEnv("DB_PORT", 5432),
		dbUser:                  LookupEnv("DB_USER", "postgres"),
		dbPassword:              LookupEnv("DB_PASSWORD", ""),
		dbConnectionTimeout:     LookupEnv("DB_CONNECT_TIMEOUT", 5),
		dbPoolMinConnections:    LookupEnv("DB_POOL_MIN_CONNS", 0),
		dbPoolMaxConnections:    LookupEnv("DB_POOL_MAX_CONNS", 4),
		dbPoolMaxConnLifetime:   LookupEnv("DB_POOL_MAX_CONN_LIFETIME", time.Hour),
		dbPoolMaxConnIdleTime:   LookupEnv("DB_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute),
		dbPoolHealthCheckPeriod: LookupEnv("DB_POOL_HEALTH_CHECK_PERIOD", time.Minute),

		jwtSecretKey: LookupEnv("JWT_SECRET_KEY", "secret"),
		jwtExpiresIn: LookupEnv("JWT_EXPIRES_IN", time.Hour*24),

		llmProvider: LookupEnv("LLM_PROVIDER", "openai"),
		llmAPIKey:   LookupEnv("LLM_API_KEY", ""),
		llmModel:    LookupEnv("LLM_MODEL", "gpt-4o-mini"),

		googleClientID:     LookupEnv("GOOGLE_CLIENT_ID", ""),
		googleClientSecret: LookupEnv("GOOGLE_CLIENT_SECRET", ""),

		otlpExporterEndpoint: LookupEnv("OTLP_EXPORTER_ENDPOINT", "localhost:4317"),
	}
}

func AppName() string     { return config.appName }
func AppEnv() string      { return config.appEnv }
func IsProduction() bool  { return config.appEnv == "production" }
func IsDevelopment() bool { return config.appEnv != "production" }
func Port() string        { return config.port }

func DBName() string                         { return config.dbName }
func DBHost() string                         { return config.dbHost }
func DBPort() int                            { return config.dbPort }
func DBUser() string                         { return config.dbUser }
func DBPassword() string                     { return config.dbPassword }
func DBConnectionTimeout() int               { return config.dbConnectionTimeout }
func DBPoolMaxConnections() int              { return config.dbPoolMaxConnections }
func DBPoolMinConnections() int              { return config.dbPoolMinConnections }
func DBPoolMaxConnLifetime() time.Duration   { return config.dbPoolMaxConnLifetime }
func DBPoolMaxConnIdleTime() time.Duration   { return config.dbPoolMaxConnIdleTime }
func DBPoolHealthCheckPeriod() time.Duration { return config.dbPoolHealthCheckPeriod }

func JWTSecretKey() string        { return config.jwtSecretKey }
func JWTExpiresIn() time.Duration { return config.jwtExpiresIn }

func SetJWTSecretKey(key string)             { config.jwtSecretKey = key }
func SetJWTExpiresIn(duration time.Duration) { config.jwtExpiresIn = duration }

func LLMProvider() llm.Provider { return llm.Provider(config.llmProvider) }

func LLMAPIKey() string { return config.llmAPIKey }
func LLMModel() string  { return config.llmModel }

func GoogleClientID() string     { return config.googleClientID }
func GoogleClientSecret() string { return config.googleClientSecret }

func OTLPExporterEndpoint() string { return config.otlpExporterEndpoint }
