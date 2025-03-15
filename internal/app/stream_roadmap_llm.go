package app

import (
	"context"
	"strings"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/pkg/llm"
	"go.opentelemetry.io/otel/attribute"
)

func (app *application) StreamRoadmapLLM(ctx context.Context, input io.StreamRoadmapLLMInput) (llm.Stream, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.StreamRoadmapLLM)")
	defer span.End()

	span.SetAttributes(
		attribute.String("roadmap_title", input.Title),
		attribute.String("user_message", input.Message))

	// TODO: store each chat session in the database
	return app.llm.Stream(ctx, llm.ChatPrompt{
		System: app.makeSystemPrompt(input),
		User:   input.Message,
	})
}

func (app *application) makeSystemPrompt(input io.StreamRoadmapLLMInput) string {
	var sb strings.Builder

	sb.WriteString("You are an AI assistant helping a user navigate the ")
	sb.WriteString(input.Title)
	sb.WriteString(" roadmap. The user will ask questions about the roadmap, and you will provide answers.\n\n")

	sb.WriteString("Here is the roadmap details:\n")
	sb.WriteString("Title: ")
	sb.WriteString(input.Title)
	sb.WriteString("\n")
	sb.WriteString("Description: ")
	sb.WriteString(input.Description)
	sb.WriteString("\n\n")

	sb.WriteString("Here are the topics and subtopics in the roadmap:\n")
	for _, topic := range input.Topics {
		sb.WriteString("- ")
		sb.WriteString(topic.Title)
		sb.WriteString(": ")
		sb.WriteString(topic.Description)
		sb.WriteString("\n")
		for _, subtopic := range topic.Subtopics {
			sb.WriteString("  - ")
			sb.WriteString(subtopic.Title)
			sb.WriteString(": ")
			sb.WriteString(subtopic.Description)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Your answer should be relevant to the roadmap and the user's question. Your output format must be in plain text.\n\n")

	return sb.String()
}
