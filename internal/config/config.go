package config

import (
	"time"

	"github.com/roadmap-thesis/backend/internal/env"
	"github.com/roadmap-thesis/backend/internal/llm"
)

// Config is the global config for the application.
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

	redisDB       int
	redisNetwork  string
	redisAddr     string
	redisUsername string
	redisPassword string

	accessSecretKey  string
	accessExpiresIn  time.Duration
	refreshSecretKey string
	refreshExpiresIn time.Duration

	llmProvider string
	llmAPIKey   string
	llmModel    string

	googleClientID     string
	googleClientSecret string

	googleBooksAPIKey string
	youtubeAPIKey     string

	otlpExporterEndpoint string
}

var config *Config

// Init initializes the config package.
func Init() {
	config = &Config{
		appName: env.LookupEnv("APP_NAME", "roadmap_application"),
		appEnv:  env.LookupEnv("APP_ENV", "local"),
		port:    env.LookupEnv("PORT", "5000"),

		dbName:                  env.LookupEnv("DB_NAME", "roadmap"),
		dbHost:                  env.LookupEnv("DB_HOST", "localhost"),
		dbPort:                  env.LookupEnv("DB_PORT", 5432),
		dbUser:                  env.LookupEnv("DB_USER", "postgres"),
		dbPassword:              env.LookupEnv("DB_PASSWORD", ""),
		dbConnectionTimeout:     env.LookupEnv("DB_CONNECT_TIMEOUT", 5),
		dbPoolMinConnections:    env.LookupEnv("DB_POOL_MIN_CONNS", 0),
		dbPoolMaxConnections:    env.LookupEnv("DB_POOL_MAX_CONNS", 4),
		dbPoolMaxConnLifetime:   env.LookupEnv("DB_POOL_MAX_CONN_LIFETIME", time.Hour),
		dbPoolMaxConnIdleTime:   env.LookupEnv("DB_POOL_MAX_CONN_IDLE_TIME", 30*time.Minute),
		dbPoolHealthCheckPeriod: env.LookupEnv("DB_POOL_HEALTH_CHECK_PERIOD", time.Minute),

		redisDB:       env.LookupEnv("REDIS_DB", 0),
		redisNetwork:  env.LookupEnv("REDIS_NETWORK", "tcp"),
		redisAddr:     env.LookupEnv("REDIS_ADDR", "localhost:6379"),
		redisUsername: env.LookupEnv("REDIS_USERNAME", ""),
		redisPassword: env.LookupEnv("REDIS_PASSWORD", ""),

		accessSecretKey:  env.LookupEnv("ACCESS_SECRET_KEY", "secret"),
		accessExpiresIn:  env.LookupEnv("ACCESS_EXPIRES_IN", time.Minute*5),
		refreshSecretKey: env.LookupEnv("REFRESH_SECRET_KEY", "secret"),
		refreshExpiresIn: env.LookupEnv("REFRESH_EXPIRES_IN", time.Hour*24*30),

		llmProvider: env.LookupEnv("LLM_PROVIDER", "openai"),
		llmAPIKey:   env.LookupEnv("LLM_API_KEY", ""),
		llmModel:    env.LookupEnv("LLM_MODEL", "gpt-4o-mini"),

		googleClientID:     env.LookupEnv("GOOGLE_CLIENT_ID", ""),
		googleClientSecret: env.LookupEnv("GOOGLE_CLIENT_SECRET", ""),

		googleBooksAPIKey: env.LookupEnv("GOOGLE_BOOKS_API_KEY", ""),
		youtubeAPIKey:     env.LookupEnv("YOUTUBE_API_KEY", ""),

		otlpExporterEndpoint: env.LookupEnv("OTLP_EXPORTER_ENDPOINT", "localhost:4317"),
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

func RedisDB() int          { return config.redisDB }
func RedisNetwork() string  { return config.redisNetwork }
func RedisAddr() string     { return config.redisAddr }
func RedisUsername() string { return config.redisUsername }
func RedisPassword() string { return config.redisPassword }

func AccessSecretKey() string         { return config.accessSecretKey }
func AccessExpiresIn() time.Duration  { return config.accessExpiresIn }
func RefreshSecretKey() string        { return config.refreshSecretKey }
func RefreshExpiresIn() time.Duration { return config.refreshExpiresIn }

func SetAccessSecretKey(key string)              { config.accessSecretKey = key }
func SetAccessExpiresIn(duration time.Duration)  { config.accessExpiresIn = duration }
func SetRefreshSecretKey(key string)             { config.refreshSecretKey = key }
func SetRefreshExpiresIn(duration time.Duration) { config.refreshExpiresIn = duration }

func LLMProvider() llm.Provider { return llm.Provider(config.llmProvider) }

func LLMAPIKey() string { return config.llmAPIKey }
func LLMModel() string  { return config.llmModel }

func GoogleClientID() string     { return config.googleClientID }
func GoogleClientSecret() string { return config.googleClientSecret }

func GoogleBooksAPIKey() string { return config.googleBooksAPIKey }
func YoutubeAPIKey() string     { return config.youtubeAPIKey }

func OTLPExporterEndpoint() string { return config.otlpExporterEndpoint }
