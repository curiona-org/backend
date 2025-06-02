package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/interval"
	"github.com/curiona-org/backend/pkg/llm"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) RegenerateRoadmap(ctx context.Context, input io.RegenerateRoadmapInput) (io.RegenerateRoadmapOutput, error) {
	traceCtx, span := app.tracer.Start(ctx, "(*application.RegenerateRoadmap)", trace.WithAttributes(
		attribute.String("slug", input.Slug),
		attribute.String("reason", input.Reason),
	))
	defer span.End()

	var output io.RegenerateRoadmapOutput

	baseRoadmap, err := app.repository.Roadmap.GetBySlug(ctx, filter.Filters{
		Slug: input.Slug,
	})
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return io.RegenerateRoadmapOutput{}, cerrors.ErrNotFound.Msg("roadmap")
		}
		return io.RegenerateRoadmapOutput{}, err
	}

	systemPrompt := app.makeRegenerateRoadmapSystemPrompt(traceCtx, baseRoadmap)
	if systemPrompt == "" {
		return io.RegenerateRoadmapOutput{}, cerrors.ErrLLMPromptGenerationFailed
	}

	generated, err := app.chatRegeneratePrompt(traceCtx, llm.ChatPrompt{
		System: systemPrompt,
		User:   app.makeRegenerateRoadmapUserPrompt(input),
	})
	if err != nil {
		return io.RegenerateRoadmapOutput{}, err
	}

	roadmap := domain.NewRoadmap(input.AccountID, generated.Title, generated.Description)

	for _, topic := range generated.Topics {
		newTopic := domain.NewTopic(input.AccountID, topic.Title, topic.Description, topic.ProTips, topic.SearchQuery)
		roadmap.TotalTopics++
		roadmap.AddTopic(newTopic)
		if len(topic.Subtopics) > 0 {
			for _, subtopic := range topic.Subtopics {
				newSubtopic := domain.NewTopic(input.AccountID, subtopic.Title, subtopic.Description, subtopic.ProTips, subtopic.SearchQuery)
				roadmap.TotalTopics++
				newTopic.AddSubtopic(newSubtopic)
			}
		}
	}

	personalizationOpt := domain.NewPersonalizationOptions(
		input.AccountID,
		0,
		interval.New(
			input.PersonalizationOptions.DailyTimeAvailability.Value,
			input.PersonalizationOptions.DailyTimeAvailability.Unit,
		).Duration(),
		interval.New(
			input.PersonalizationOptions.TotalDuration.Value,
			input.PersonalizationOptions.TotalDuration.Unit,
		).Duration(),
		domain.SkillLevel(input.PersonalizationOptions.SkillLevel),
		input.PersonalizationOptions.AdditionalInfo,
	)
	roadmap.SetPersonalizationOptions(personalizationOpt)

	createdRoadmap, err := app.repository.Roadmap.Save(traceCtx, roadmap)
	if err != nil {
		return io.RegenerateRoadmapOutput{}, err
	}

	output.Slug = createdRoadmap.Slug

	return output, nil
}

type chatRegeneratePromptPromptResult struct {
	Title       string                                  `json:"title"`
	Description string                                  `json:"description"`
	Topics      []chatRegeneratePromptPromptResultTopic `json:"topics"`
}

type chatRegeneratePromptPromptResultTopic struct {
	Title       string                                     `json:"title"`
	Description string                                     `json:"description"`
	ProTips     string                                     `json:"pro_tips"`
	Subtopics   []chatRegeneratePromptPromptResultSubtopic `json:"subtopics"`
	SearchQuery string                                     `json:"search_query"`
}

type chatRegeneratePromptPromptResultSubtopic struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ProTips     string `json:"pro_tips"`
	SearchQuery string `json:"search_query"`
}

func (app *application) chatRegeneratePrompt(ctx context.Context, prompt llm.ChatPrompt) (chatRegeneratePromptPromptResult, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.chatRegeneratePrompt)")
	defer span.End()

	span.SetAttributes(
		attribute.String("prompt.system", prompt.System),
		attribute.String("prompt.user", prompt.User),
	)

	content, err := app.llm.Chat(ctx, prompt)
	if err != nil {
		span.RecordError(err)
		return chatRegeneratePromptPromptResult{}, cerrors.ErrLLMProviderUnavailable.With(err)
	}

	var result chatRegeneratePromptPromptResult

	if err = json.Unmarshal([]byte(content), &result); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return chatRegeneratePromptPromptResult{}, cerrors.ErrLLMInvalidData.With(err)
	}

	return result, nil
}

func (app *application) makeRegenerateRoadmapUserPrompt(input io.RegenerateRoadmapInput) string {
	var sb strings.Builder
	sb.WriteString(`I will give you a topic and you need to generate a learning roadmap for it. Just reply to the question without adding any other information about the prompt and use simple language.
`)

	sb.WriteString("Generate a structured learning roadmap for the topic: ")
	sb.WriteString(input.Reason)

	sb.WriteString("\nHere are my personalization options:\n")

	sb.WriteString("- Daily Time Availability: ")
	sb.WriteString(strconv.Itoa(input.PersonalizationOptions.DailyTimeAvailability.Value))
	sb.WriteString(" ")
	sb.WriteString(input.PersonalizationOptions.DailyTimeAvailability.Unit.String())
	sb.WriteString("\n")

	sb.WriteString("- Total Duration: ")
	sb.WriteString(strconv.Itoa(input.PersonalizationOptions.TotalDuration.Value))
	sb.WriteString(" ")
	sb.WriteString(input.PersonalizationOptions.TotalDuration.Unit.String())
	sb.WriteString("\n")

	sb.WriteString("- Skill Level: ")
	sb.WriteString(input.PersonalizationOptions.SkillLevel)
	sb.WriteString("\n")

	if input.PersonalizationOptions.AdditionalInfo != "" {
		sb.WriteString("- Additional Information:\n \"\"\"\n ")
		sb.WriteString(input.PersonalizationOptions.AdditionalInfo)
		sb.WriteString("\n \"\"\"\n")
	}

	return sb.String()
}

func (app *application) makeRegenerateRoadmapSystemPrompt(ctx context.Context, baseRoadmap domain.Roadmap) string {
	var sb strings.Builder

	log := logger.FromContext(ctx)

	promptUserPersonalizationOptions := []string{
		"Daily Time Availability: How much time the user can dedicate daily (e.g., 15 minutes, 30 minutes, 1 hour).",
		"Total Duration: The overall duration of the roadmap (e.g., 1 week, 3 months).",
		"Skill Level: The user's experience level (e.g., Beginner, Intermediate, Advanced).",
		"Additional Info: Any other user-provided goals or preferences. This is Optional for the user.",
	}

	promptSystemGuidelines := []string{
		"Go into detail about the main topic to provide a comprehensive overview of the subject.",
		"Each topic should have a title, a brief description to explain the focus of that section, and a pro tip to help the user understand the topic better.",
		"Subtopics should be related to the main topic and provide more detailed information on specific aspects of the subject.",
		"Each description should be clear and informative. It should be long enough to explain the topic but concise enough to maintain the user's interest.",
		"Pro tips should be practical and relevant to the topic, providing additional insights or shortcuts to help the user learn more effectively.",
		"Ensure that a topic is broken down into manageable subtopics to help users understand the subject better whenever possible.",
		"A topic can also not have any subtopics if it is a standalone subject.",
		"Use only English language for the roadmap.",
		fmt.Sprintf("Must have a minimum of %d topics and %d (or more) subtopics per topic.", domain.RoadmapMinimumTopics, domain.RoadmapMinimumSubtopics),
		fmt.Sprintf("Must have a maximum of %d topics and %d (or less) subtopics per topic.", domain.RoadmapMaximumTopics, domain.RoadmapMaximumSubtopics),
		"Each topic and subtopic should have a search query that can be used to find more information on the topic online",
		"Make sure the search query is relevant to the topic and provides accurate results as it will be used by the system to fetch books, youtube videos, and other resources.",
		"If for example the topic of learning golang be \"Introduction\" make the search query \"Introduction Golang\".",
	}

	exampleFormat := chatRegeneratePromptPromptResult{
		Title:       "Example Topic",
		Description: "An extensive overview of the topic to set the stage for learning.",
		Topics: []chatRegeneratePromptPromptResultTopic{
			{
				Title:       "Main Topic",
				Description: "A one paragraph long explanation of the main topic.",
				ProTips:     "Some pro tips to help the user understand the topic better.",
				SearchQuery: "Main Topic",
				Subtopics: []chatRegeneratePromptPromptResultSubtopic{
					{
						Title:       "Subtopic 1",
						Description: "A one paragraph long explanation of Subtopic 1.",
						ProTips:     "Some pro tips to help the user understand subtopic 1 better.",
						SearchQuery: "Subtopic 1",
					},
					{
						Title:       "Subtopic 2",
						Description: "A one paragraph long explanation of Subtopic 2.",
						ProTips:     "Some pro tips to help the user understand subtopic 2 better.",
						SearchQuery: "Subtopic 2",
					},
				},
			},
		},
	}

	baseRoadmapPrompt := chatRegeneratePromptPromptResult{
		Title:       baseRoadmap.Title,
		Description: baseRoadmap.Description,
		Topics:      make([]chatRegeneratePromptPromptResultTopic, 0, len(baseRoadmap.Topics)),
	}

	for _, topic := range baseRoadmap.Topics {
		newTopic := chatRegeneratePromptPromptResultTopic{
			Title:       topic.Title,
			Description: topic.Description,
			ProTips:     topic.ProTips,
			SearchQuery: topic.ExternalSearchQuery,
			Subtopics:   make([]chatRegeneratePromptPromptResultSubtopic, 0, len(topic.Subtopics)),
		}
		for _, subtopic := range topic.Subtopics {
			newSubtopic := chatRegeneratePromptPromptResultSubtopic{
				Title:       subtopic.Title,
				Description: subtopic.Description,
				ProTips:     subtopic.ProTips,
				SearchQuery: subtopic.ExternalSearchQuery,
			}
			newTopic.Subtopics = append(newTopic.Subtopics, newSubtopic)
		}
		baseRoadmapPrompt.Topics = append(baseRoadmapPrompt.Topics, newTopic)
	}

	sb.WriteString(`You are an expert in refining existing learning roadmaps based on the given roadmap the user needed revision on, user feedback and new requirements. Your task is to modify an existing roadmap according to the provided regeneration request while maintaining all original structural requirements. The regeneration process will:
1. Use the provided roadmap as a base, because the user wants to regenerate it, not create a new one from scratch. Here is the base roadmap:
`)

	baseRoadmapJSON, err := json.MarshalIndent(baseRoadmapPrompt, "", "    ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal base roadmap")
		return ""
	}
	sb.Write(baseRoadmapJSON)

	sb.WriteString("\n")
	sb.WriteString(`
2. Analyze the user's stated reason for regeneration (e.g., difficulty mismatch, time constraints, content preferences)
3. Be tailored based on user-provided personalization options:
`)

	for _, userPersonalizationOpt := range promptUserPersonalizationOptions {
		sb.WriteString(" - ")
		sb.WriteString(userPersonalizationOpt)
		sb.WriteString("\n")
	}

	sb.WriteString(`
3. Preserve all original roadmap requirements:
 - 5-10 topics with 3-5 subtopics each
 - English-only content
 - Search queries for all entries
 - Pro tips in every section
4. Make targeted adjustments to:
 - Topic/subtopic selection and depth
 - Time allocation per section
 - Difficulty progression
 - Content focus areas
 - Resource search queries

# Regeneration Guidelines:
1. Start by analyzing the user's reason to identify required changes
2. Adjust timeline distribution based on new daily/time total constraints
3. Modify complexity based on updated skill level (Beginner/Intermediate/Advanced)
4. Add/remove topics based on user feedback and personalization options
5. Maintain JSON structure integrity throughout modifications
6. Preserve valuable content from original roadmap where appropriate
7. Ensure search queries remain relevant to modified content
8. Update pro tips to match adjusted difficulty level
9. Verify all descriptions remain clear and concise
10. Strictly follow original format requirements for subtopic nesting`)

	sb.WriteString("# Guidelines:\n")
	for _, guideline := range promptSystemGuidelines {
		sb.WriteString(" - ")
		sb.WriteString(guideline)
		sb.WriteString("\n")
	}

	sb.WriteString("# Example Format:\n")

	exampleFormatJSON, err := json.MarshalIndent(exampleFormat, "", "    ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal example format")
		return ""
	}

	sb.Write(exampleFormatJSON)

	sb.WriteString("\n")
	sb.WriteString("The roadmap must adhere to this format while reflecting the user's provided topic and personalization preferences. Do not use markdown symbols such as the triple backticks or quotes, you must only respond with the raw json itself.")
	sb.WriteString("\n")

	return sb.String()
}
