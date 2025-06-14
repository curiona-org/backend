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

	// Validate if user already hit the limit of generating roadmaps
	account, err := app.repository.Account.GetByID(ctx, input.AccountID)
	if err != nil {
		return io.RegenerateRoadmapOutput{}, cerrors.ErrUnauthorized
	}

	if account.IsSuspended {
		return io.RegenerateRoadmapOutput{}, cerrors.ErrForbidden
	}

	if !account.IsAdmin {
		// Check if the account has reached the maximum number of generated roadmaps by
		// checking the number of unfinished roadmaps.
		accountRoadmapsCount, err := app.repository.Roadmap.CountUnfinishedRoadmapsByAccountID(ctx, input.AccountID)
		if err != nil {
			return io.RegenerateRoadmapOutput{}, err
		}

		if accountRoadmapsCount >= uint64(account.Profile.MaxGeneratedRoadmaps) {
			return io.RegenerateRoadmapOutput{}, cerrors.ErrLLMMaximumRoadmapGenerationReached
		}
	}

	flagged, err := app.llm.Moderate(traceCtx, input.Reason)
	if err != nil {
		return io.RegenerateRoadmapOutput{}, err
	}

	if flagged {
		return io.RegenerateRoadmapOutput{
			Flagged: true,
			Reason:  cerrors.ErrLLMFlaggedContentDetected.Message(),
		}, nil
	}

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
		User:   app.makeRegenerateRoadmapUserPrompt(baseRoadmap.Title, input),
	})
	if err != nil {
		return io.RegenerateRoadmapOutput{}, err
	}

	if generated.Flagged {
		return io.RegenerateRoadmapOutput{
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
		return io.RegenerateRoadmapOutput{}, err
	}

	output.Slug = createdRoadmap.Slug

	return output, nil
}

type chatRegeneratePromptPromptResult struct {
	Flagged       bool                                    `json:"flagged"`
	FlaggedReason string                                  `json:"flagged_reason"`
	Title         string                                  `json:"title"`
	Description   string                                  `json:"description"`
	Topics        []chatRegeneratePromptPromptResultTopic `json:"topics"`
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

func (app *application) makeRegenerateRoadmapUserPrompt(roadmapTitle string, input io.RegenerateRoadmapInput) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`I'd like to refine my existing learning roadmap for "%s". Here's why I need changes:\n`, roadmapTitle))

	sb.WriteString("Reason for regeneration: ")
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

	sb.WriteString("Please adjust the roadmap to better match my needs while keeping the same overall structure. Thank you!\n")

	return sb.String()
}

func (app *application) makeRegenerateRoadmapSystemPrompt(ctx context.Context, baseRoadmap domain.Roadmap) string {
	var sb strings.Builder

	log := logger.FromContext(ctx)

	baseRoadmapPrompt := chatRegeneratePromptPromptResult{
		Flagged:       false,
		FlaggedReason: "",
		Title:         baseRoadmap.Title,
		Description:   baseRoadmap.Description,
		Topics:        make([]chatRegeneratePromptPromptResultTopic, 0, len(baseRoadmap.Topics)),
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

	sb.WriteString(`You are an expert in refining learning roadmaps. Your task is to modify the existing roadmap based on user feedback and requirements while preserving the original structure.\n\n`)

	sb.WriteString(`Here is the base roadmap:\n`)

	baseRoadmapJSON, err := json.MarshalIndent(baseRoadmapPrompt, "", "    ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal base roadmap")
		return ""
	}
	sb.Write(baseRoadmapJSON)

	sb.WriteString("\n")
	sb.WriteString("INPUT:\n")
	sb.WriteString("1. Base roadmap (JSON): The existing roadmap structure the user wants to modify\n")
	sb.WriteString("2. User's reason for regeneration: Difficulty mismatch, time constraints, etc.\n")
	sb.WriteString("3. If the user provided reason contains any sensitive or inappropriate content, it should set flagged into true, provide an accurate & concise but brief & short reasoning why it's flagged and not generate a new roadmap.\n")
	sb.WriteString("4. Personalization options:\n")
	sb.WriteString("   - Daily Time Availability (e.g., 15 min, 30 min, 1 hour)\n")
	sb.WriteString("   - Total Duration (e.g., 1 week, 3 months)\n")
	sb.WriteString("   - Skill Level (Beginner, Intermediate, Advanced)\n")
	sb.WriteString("   - Additional Goals/Preferences (optional)\n")

	sb.WriteString("\n")
	sb.WriteString("REQUIREMENTS:\n")
	sb.WriteString("1. Preserve structural elements:\n")
	sb.WriteString("   - Search queries for all topics/subtopics\n")
	sb.WriteString("   - Pro tips in every section\n")
	sb.WriteString("   - JSON format integrity\n")
	sb.WriteString("2. Make targeted adjustments to:\n")
	sb.WriteString("   - Topic selection and depth\n")
	sb.WriteString("   - Time allocation\n")
	sb.WriteString("   - Difficulty progression\n")
	sb.WriteString("   - Content focus areas\n")

	sb.WriteString("\n")
	sb.WriteString("PROCESS:\n")
	sb.WriteString("1. Analyze the user's reason for regeneration\n")
	sb.WriteString("2. Adjust timeline based on new time constraints\n")
	sb.WriteString("3. Modify complexity to match skill level\n")
	sb.WriteString("4. Add/remove topics based on user feedback\n")
	sb.WriteString("5. Ensure search queries remain relevant\n")
	sb.WriteString("6. Update pro tips to match difficulty level\n")

	sb.WriteString("\n")
	sb.WriteString("FORMAT RULES:\n")
	sb.WriteString(fmt.Sprintf("- Maximum %d topics with up to %d subtopics each.\n", domain.RoadmapMaximumTopics, domain.RoadmapMaximumSubtopics))
	sb.WriteString("- Each topic needs: title, description, pro_tips, search_query\n")
	sb.WriteString("- Each subtopic needs: title, description, pro_tips, search_query\n")
	sb.WriteString("- Descriptions should be clear, informative paragraphs\n")
	sb.WriteString("- Pro tips should provide practical insights\n")
	sb.WriteString("- Search queries must be relevant for finding resources\n")
	sb.WriteString("- Return only raw JSON without markdown formatting\n")

	sb.WriteString("\n")
	sb.WriteString("The final output must maintain the original JSON structure while incorporating the user's personalization preferences. Do not use markdown symbols such as the triple backticks or quotes, you must only respond with the raw json itself.\n")

	return sb.String()
}
