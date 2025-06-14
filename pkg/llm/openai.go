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
		attribute.String("prompt", userPrompt),
		attribute.Bool("flagged", result.Flagged),

		attribute.Bool("category.harassment", result.Categories.Harassment),
		attribute.Float64("category.score.harassment", float64(result.CategoryScores.Harassment)),

		attribute.Bool("category.harassment_threatening", result.Categories.HarassmentThreatening),
		attribute.Float64("category.score.harassment_threatening", float64(result.CategoryScores.HarassmentThreatening)),

		attribute.Bool("category.hate", result.Categories.Hate),
		attribute.Float64("category.score.hate", float64(result.CategoryScores.Hate)),

		attribute.Bool("category.hate_threatening", result.Categories.HateThreatening),
		attribute.Float64("category.score.hate_threatening", float64(result.CategoryScores.HateThreatening)),

		attribute.Bool("category.self_harm", result.Categories.SelfHarm),
		attribute.Float64("category.score.self_harm", float64(result.CategoryScores.SelfHarm)),

		attribute.Bool("category.self_harm_instructions", result.Categories.SelfHarmInstructions),
		attribute.Float64("category.score.self_harm_instructions", float64(result.CategoryScores.SelfHarmInstructions)),

		attribute.Bool("category.self_harm_threatening", result.Categories.SelfHarmIntent),
		attribute.Float64("category.score.self_harm_threatening", float64(result.CategoryScores.SelfHarmIntent)),

		attribute.Bool("category.sexual", result.Categories.Sexual),
		attribute.Float64("category.score.sexual", float64(result.CategoryScores.Sexual)),

		attribute.Bool("category.sexual_minors", result.Categories.SexualMinors),
		attribute.Float64("category.score.sexual_minors", float64(result.CategoryScores.SexualMinors)),

		attribute.Bool("category.violence", result.Categories.Violence),
		attribute.Float64("category.score.violence", float64(result.CategoryScores.Violence)),

		attribute.Bool("category.violence_graphic", result.Categories.ViolenceGraphic),
		attribute.Float64("category.score.violence_graphic", float64(result.CategoryScores.ViolenceGraphic)),
	)

	return result.Flagged, nil
}
