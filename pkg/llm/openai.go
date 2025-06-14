package llm

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type openAiClient struct {
	client   *openai.Client
	model    string
	tracer   trace.Tracer
	provider Provider
}

var _ Client = (*openAiClient)(nil)

func NewOpenAiClient(provider Provider, authToken, model string) Client {
	cfg := openai.DefaultConfig(authToken)
	if provider == Groq {
		cfg.BaseURL = "https://api.groq.com/openai/v1"
	}

	client := openai.NewClientWithConfig(cfg)

	tracer := otel.Tracer("llm:openai")
	return &openAiClient{
		client: client,
		model:  model,
		tracer: tracer,
	}
}

func (o *openAiClient) Chat(ctx context.Context, prompt ChatPrompt) (string, error) {
	ctx, span := o.tracer.Start(ctx, "(*openAiClient.Chat)")
	defer span.End()

	response, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt.System,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt.User,
			},
		},
	})
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	span.SetAttributes(
		attribute.String("id", response.ID),
		attribute.String("model", response.Model),
		attribute.String("object", response.Object),
		attribute.Int64("created", response.Created),
		attribute.Int("completion_tokens", response.Usage.CompletionTokens),
		attribute.Int("prompt_tokens", response.Usage.PromptTokens),
		attribute.Int("total_tokens", response.Usage.TotalTokens),
		attribute.String("content", response.Choices[0].Message.Content),
	)

	if len(response.Choices) == 0 {
		return "", errors.New("openai: no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func (o *openAiClient) Stream(ctx context.Context, prompt ChatPrompt) (Stream, error) {
	ctx, span := o.tracer.Start(ctx, "(*openAiClient.Stream)")
	defer span.End()

	stream, err := o.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt.System,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt.User,
			},
		},
	})

	return NewStreamHandler(&streamHandlerConfig{
		openai: stream,
	}), err
}

func (o *openAiClient) Moderate(ctx context.Context, userPrompt string) (bool, error) {
	ctx, span := o.tracer.Start(ctx, "(*openAiClient.Moderate)")
	defer span.End()

	// Groq supports moderation but requires a different endpoint or method.
	// Since we use Groq only for development, we can skip moderation for now.
	if o.provider == Groq {
		span.SetAttributes(attribute.String("provider", "groq"))
		span.SetStatus(codes.Ok, "Skipping moderation for Groq provider")
		return true, nil
	}

	response, err := o.client.Moderations(ctx, openai.ModerationRequest{
		Input: userPrompt,
		Model: "omni-moderation-latest",
	})
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	if len(response.Results) == 0 {
		return false, errors.New("openai: no results in moderation response")
	}

	result := response.Results[0]

	span.SetAttributes(
		attribute.String("id", response.ID),
		attribute.String("model", response.Model),
		attribute.Bool("flagged", result.Flagged),
	)

	return result.Flagged, nil
}
