package llm

import (
	"context"
	"errors"
	"io"

	"github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
)

var (
	ErrInvalidProvider = errors.New("unexpected llm stream handler")
)

type Client interface {
	Chat(ctx context.Context, prompt ChatPrompt) (string, error)
	Stream(ctx context.Context, prompt ChatPrompt) (Stream, error)
	Moderate(ctx context.Context, userPrompt string) (bool, error)
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
		return nil, ErrInvalidProvider
	}

	return client, nil
}

type StreamResponse interface {
	openai.ChatCompletionStreamResponse | deepseek.StreamChatCompletionResponse
}

type Stream interface {
	Recv() (string, error)
	Close() error
}

var (
	ErrStreamDone = io.EOF
)

type stream struct {
	openai   *openai.ChatCompletionStream
	deepseek deepseek.ChatCompletionStream
}

type streamHandlerConfig struct {
	openai   *openai.ChatCompletionStream
	deepseek deepseek.ChatCompletionStream
}

func NewStreamHandler(cfg *streamHandlerConfig) Stream {
	return &stream{
		openai:   cfg.openai,
		deepseek: cfg.deepseek,
	}
}

func (s *stream) Recv() (string, error) {
	if s.openai != nil {
		recv, err := s.openai.Recv()
		if err != nil {
			return "", err
		}

		return recv.Choices[0].Delta.Content, err
	}

	if s.deepseek != nil {
		recv, err := s.deepseek.Recv()
		if err != nil {
			return "", err
		}

		return recv.Choices[0].Delta.Content, err
	}

	return "", ErrInvalidProvider
}

func (s *stream) Close() error {
	if s.openai != nil {
		return s.openai.Close()
	}

	if s.deepseek != nil {
		return s.deepseek.Close()
	}

	return ErrInvalidProvider
}
