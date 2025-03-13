package llm

import (
	"context"
	"errors"

	"github.com/cohesion-org/deepseek-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type deepSeekClient struct {
	client *deepseek.Client
	model  string
	tracer trace.Tracer
}

var _ Client = (*deepSeekClient)(nil)

func NewDeepSeekClient(authToken, model string) Client {
	client := deepseek.NewClient(authToken)
	tracer := otel.Tracer("llm:deepseek")

	return &deepSeekClient{
		client: client,
		model:  model,
		tracer: tracer,
	}
}

func (d *deepSeekClient) Chat(ctx context.Context, prompt ChatPrompt) (string, error) {
	ctx, span := d.tracer.Start(ctx, "(*deepSeekClient.Chat)")
	defer span.End()

	response, err := d.client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model: d.model,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleSystem, Content: prompt.System},
			{Role: deepseek.ChatMessageRoleUser, Content: prompt.User},
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
		return "", errors.New("deepseek: no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func (d *deepSeekClient) Stream(ctx context.Context, prompt ChatPrompt) (Stream, error) {
	ctx, span := d.tracer.Start(ctx, "(*deepSeekClient.Stream)")
	defer span.End()

	stream, err := d.client.CreateChatCompletionStream(ctx, &deepseek.StreamChatCompletionRequest{
		Model: d.model,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleSystem, Content: prompt.System},
			{Role: deepseek.ChatMessageRoleUser, Content: prompt.User},
		},
	})

	return NewStreamHandler(&streamHandlerConfig{
		deepseek: stream,
	}), err
}
