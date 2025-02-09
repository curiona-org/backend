package llm

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

type Client interface {
	Chat(ctx context.Context, prompt ChatPrompt) (string, error)
}

type Provider string

const (
	OpenAI   Provider = "openai"
	DeepSeek Provider = "deepseek"
)

var (
	tracer = otel.Tracer("LLM")
)

func NewClient(provider Provider, authToken, model string) (Client, error) {
	var client Client
	switch provider {
	case OpenAI:
		log.Info().Msg("using OpenAI LLM provider")
		client = NewOpenAiClient(authToken, model)
	case DeepSeek:
		log.Info().Msg("using DeepSeek LLM provider")
		client = NewDeepSeekClient(authToken, model)
	default:
		return nil, errors.New("invalid LLM provider")
	}

	return client, nil
}
