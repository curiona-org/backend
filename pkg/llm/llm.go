package llm

import (
	"context"
	"errors"
)

type Client interface {
	Chat(ctx context.Context, prompt ChatPrompt) (string, error)
}

type Provider string

const (
	OpenAI   Provider = "openai"
	DeepSeek Provider = "deepseek"
	Groq     Provider = "groq"
)

func NewClient(provider Provider, authToken, model string) (Client, error) {
	var client Client
	switch provider {
	case OpenAI, Groq:
		client = NewOpenAiClient(provider, authToken, model)
	case DeepSeek:
		client = NewDeepSeekClient(authToken, model)
	default:
		return nil, errors.New("invalid LLM provider")
	}

	return client, nil
}
