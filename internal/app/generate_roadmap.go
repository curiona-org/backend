package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/interval"
	"github.com/curiona-org/backend/pkg/llm"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) GenerateRoadmap(ctx context.Context, input io.GenerateRoadmapInput) (io.GenerateRoadmapOutput, error) {
	traceCtx, span := app.tracer.Start(ctx, "(*application.GenerateRoadmap)", trace.WithAttributes(
		attribute.String("topic", input.Topic),
		attribute.Int("personalization_options.daily_time_availability.value", input.PersonalizationOptions.DailyTimeAvailability.Value),
		attribute.String("personalization_options.daily_time_availability.unit", input.PersonalizationOptions.DailyTimeAvailability.Unit.String()),
		attribute.Int("personalization_options.total_duration.value", input.PersonalizationOptions.TotalDuration.Value),
		attribute.String("personalization_options.total_duration.unit", input.PersonalizationOptions.TotalDuration.Unit.String()),
		attribute.String("skill_level", input.PersonalizationOptions.SkillLevel),
		attribute.String("additional_info", input.PersonalizationOptions.AdditionalInfo),
	))
	defer span.End()

	// Validate if user already hit the limit of generating roadmaps or suspended
	account, err := app.repository.Account.GetByID(ctx, input.AccountID)
	if err != nil {
		return io.GenerateRoadmapOutput{}, cerrors.ErrUnauthorized
	}

	if account.IsSuspended {
		return io.GenerateRoadmapOutput{}, cerrors.ErrForbidden
	}

	if !account.IsAdmin {
		// Check if the account has reached the maximum number of generated roadmaps by
		// checking the number of unfinished roadmaps.
		accountRoadmapsCount, err := app.repository.Roadmap.CountUnfinishedRoadmapsByAccountID(ctx, input.AccountID)
		if err != nil {
			return io.GenerateRoadmapOutput{}, err
		}

		if accountRoadmapsCount >= uint64(account.Profile.MaxGeneratedRoadmaps) {
			return io.GenerateRoadmapOutput{}, cerrors.ErrLLMMaximumRoadmapGenerationReached
		}
	}

	flagged, err := app.llm.Moderate(traceCtx, input.Topic)
	if err != nil {
		return io.GenerateRoadmapOutput{}, err
	}

	if flagged {
		return io.GenerateRoadmapOutput{
			Flagged: true,
			Reason:  cerrors.ErrLLMFlaggedContentDetected.Message(),
		}, nil
	}

	var output io.GenerateRoadmapOutput

	systemPrompt := app.makeGenerateRoadmapSystemPrompt(traceCtx)
	if systemPrompt == "" {
		return io.GenerateRoadmapOutput{}, cerrors.ErrLLMPromptGenerationFailed
	}

	generated, err := app.chatGeneratePrompt(traceCtx, llm.ChatPrompt{
		System: systemPrompt,
		User:   app.makeGenerateRoadmapUserPrompt(input),
	})
	if err != nil {
		return io.GenerateRoadmapOutput{}, err
	}

	// If the moderation API considered the generated content not harmful, but the completion API
	// flagged it, we still return the flagged response.
	if generated.Flagged {
		return io.GenerateRoadmapOutput{
			Flagged: true,
			Reason:  generated.FlaggedReason,
		}, nil
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
		return io.GenerateRoadmapOutput{}, err
	}

	output.Slug = createdRoadmap.Slug

	return output, nil
}

type chatGeneratePromptPromptResult struct {
	Flagged       bool                                  `json:"flagged"`
	FlaggedReason string                                `json:"flagged_reason"`
	Title         string                                `json:"title"`
	Description   string                                `json:"description"`
	Topics        []chatGeneratePromptPromptResultTopic `json:"topics"`
}

type chatGeneratePromptPromptResultTopic struct {
	Title       string                                   `json:"title"`
	Description string                                   `json:"description"`
	ProTips     string                                   `json:"pro_tips"`
	Subtopics   []chatGeneratePromptPromptResultSubtopic `json:"subtopics"`
	SearchQuery string                                   `json:"search_query"`
}

type chatGeneratePromptPromptResultSubtopic struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ProTips     string `json:"pro_tips"`
	SearchQuery string `json:"search_query"`
}

func (app *application) chatGeneratePrompt(ctx context.Context, prompt llm.ChatPrompt) (chatGeneratePromptPromptResult, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.chatGeneratePrompt)")
	defer span.End()

	span.SetAttributes(
		attribute.String("prompt.system", prompt.System),
		attribute.String("prompt.user", prompt.User),
	)

	content, err := app.llm.Chat(ctx, prompt)
	if err != nil {
		span.RecordError(err)
		return chatGeneratePromptPromptResult{}, cerrors.ErrLLMProviderUnavailable.With(err)
	}

	var result chatGeneratePromptPromptResult

	if err = json.Unmarshal([]byte(content), &result); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return chatGeneratePromptPromptResult{}, cerrors.ErrLLMInvalidData.With(err)
	}

	return result, nil
}

func (app *application) makeGenerateRoadmapUserPrompt(input io.GenerateRoadmapInput) string {
	var sb strings.Builder
	sb.WriteString(`I will give you a topic and you need to generate a learning roadmap for it. Just reply to the question without adding any other information about the prompt and use simple language.
`)

	sb.WriteString("Generate a structured learning roadmap for the topic: ")
	sb.WriteString(input.Topic)

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

func (app *application) makeGenerateRoadmapSystemPrompt(ctx context.Context) string {
	var sb strings.Builder

	log := logger.FromContext(ctx)

	sb.WriteString("You are an expert in creating structured learning roadmaps. Your task is to generate a detailed, well-structured roadmap in JSON format based on user input, adhering to the specified guidelines and format.\n")

	sb.WriteString("\n")
	sb.WriteString("INPUT:\n")
	sb.WriteString("1. Include a title and description of the main topic to introduce the subject.\n")
	sb.WriteString("2. Break down the topic into topics and subtopics, each with a title and description to explain the focus of the section.\n")
	sb.WriteString("3. Use a maximum of 2 levels of depth for subtopics. Topics can contain subtopics, but subtopics cannot have further nested levels.\n")
	sb.WriteString("4. Be tailored based on user-provided personalization options:\n")
	sb.WriteString("   - Daily Time Availability: How much time the user can dedicate daily (e.g., 15 minutes, 30 minutes, 1 hour).\n")
	sb.WriteString("   - Total Duration: The overall duration of the roadmap (e.g., 1 week, 3 months).\n")
	sb.WriteString("   - Skill Level: The user's experience level (e.g., Beginner, Intermediate, Advanced).\n")
	sb.WriteString("   - Additional Info: Any other user-provided goals or preferences. This is Optional for the user.\n")

	sb.WriteString("\n")
	sb.WriteString("REQUIREMENTS:\n")
	sb.WriteString("1. Go into detail about the main topic to provide a comprehensive overview of the subject.\n")
	sb.WriteString("2. Each topic should have a title, a brief description to explain the focus of that section, and a pro tip to help the user understand the topic better.\n")
	sb.WriteString("3. Subtopics should be related to the main topic and provide more detailed information on specific aspects of the subject.\n")
	sb.WriteString("4. Descriptions should be clear and informative.\n")
	sb.WriteString("5. Pro tips should be practical and relevant to the topic, providing additional insights or shortcuts to help the user learn more effectively.\n")
	sb.WriteString("6. Ensure that a topic is broken down into manageable subtopics to help users understand the subject better whenever possible.\n")
	sb.WriteString("7. A topic can also not have any subtopics if it is a standalone subject.\n")
	sb.WriteString("8. Use only English language for the roadmap.\n")
	sb.WriteString(fmt.Sprintf("9. Limit to maximum %d topics; each with up to %d subtopics\n", domain.RoadmapMaximumTopics, domain.RoadmapMaximumSubtopics))
	sb.WriteString("10. Each topic and subtopic should have a search query that can be used to find more information on the topic online\n")
	sb.WriteString("11. Make sure the search query is relevant to the topic and provides accurate results as it will be used by the system to fetch books, youtube videos, and other resources.\n")
	sb.WriteString("12. If for example the topic of learning golang be \"Introduction\" make the search query something like \"Introduction Golang\".\n")
	sb.WriteString("13. Return only raw JSON, with no markdown or quotes\n")

	sb.WriteString("\n")
	sb.WriteString("INAPPROPRIATE CONTENT GUIDELINES:\n")
	sb.WriteString("1. The system should flag the content if it contains any inappropriate content, in the following categories:\n")
	sb.WriteString("   - Violence or Harmful Activities: Topics that promote violence, self-harm, or illegal activities.\n")
	sb.WriteString("   - Hate Speech or Discrimination: Topics that promote hate speech, discrimination, or intolerance against individuals or groups.\n")
	sb.WriteString("   - Adult Content: Topics that contain explicit or adult content, including sexually explicit material and or exploitation of minors.\n")
	sb.WriteString("   - Self Harm: Topics that promote self with intent or threat to self with intent.\n")
	sb.WriteString("   - Illegal Activities: Topics that promote illegal activities, such as drug manufacturing, hacking (e.g., how to hack a bank), or other criminal activities.\n")
	sb.WriteString("2. If the topic is historical, political, or educational in nature, it should not be flagged\n")
	sb.WriteString("3. If the topic is flagged, the system should flag the content and return a reason for the flagging.\n")

	sb.WriteString("\n")
	sb.WriteString("FORMAT:\n")
	sb.WriteString("The roadmap must adhere to this format while reflecting the user's provided topic and personalization preferences. Do not use markdown symbols such as the triple backticks or quotes, you must only respond with the raw json itself.\n")

	exampleFormat := chatGeneratePromptPromptResult{
		Flagged:       false,
		FlaggedReason: "",
		Title:         "Example Topic",
		Description:   "An extensive overview of the topic to set the stage for learning.",
		Topics: []chatGeneratePromptPromptResultTopic{
			{
				Title:       "Main Topic",
				Description: "A one paragraph long explanation of the main topic.",
				ProTips:     "Some pro tips to help the user understand the topic better.",
				SearchQuery: "Main Topic",
				Subtopics: []chatGeneratePromptPromptResultSubtopic{
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
	exampleFormatJSON, err := json.MarshalIndent(exampleFormat, "", "    ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal example format")
		return ""
	}

	sb.Write(exampleFormatJSON)

	return sb.String()
}
