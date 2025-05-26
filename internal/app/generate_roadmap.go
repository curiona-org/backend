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
	Title       string                                `json:"title"`
	Description string                                `json:"description"`
	Topics      []chatGeneratePromptPromptResultTopic `json:"topics"`
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

	exampleFormat := chatGeneratePromptPromptResult{
		Title:       "Example Topic",
		Description: "An extensive overview of the topic to set the stage for learning.",
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

	exampleResult := chatGeneratePromptPromptResult{
		Title:       "Front End Development",
		Description: "Step by step guide to learn  frontend development.",
		Topics: []chatGeneratePromptPromptResultTopic{
			{
				Title:       "What Is Front End Dev?",
				Description: "Front end development is the practice of producing HTML, CSS, and JavaScript for a website or web application so a user can see and interact with them directly. It involves the design of the site, the layout, the colors, the fonts, and so on.",
				ProTips:     "Focus on the basics of HTML, CSS, and JavaScript first.",
				SearchQuery: "Front End Development",
				Subtopics: []chatGeneratePromptPromptResultSubtopic{
					{
						Title:       "HTML",
						Description: "HTML is the standard markup language for creating web pages and web applications. It provides the basic structure of sites, which is enhanced and modified by other technologies like CSS and JavaScript.",
						ProTips:     "Try to understand the semantic structure of HTML.",
						SearchQuery: "HTML",
					},
					{
						Title:       "CSS",
						Description: "CSS is a style sheet language used for describing the presentation of a document written in HTML. It controls the layout of multiple web pages all at once.",
						ProTips:     "Learn about Flexbox and Grid for layout.",
						SearchQuery: "CSS",
					},
					{
						Title:       "JavaScript",
						Description: "JavaScript is a programming language that enables you to interact with elements on a webpage. It is used for creating dynamic and interactive web pages.",
						ProTips:     "Start with the basics of JavaScript syntax and DOM manipulation.",
						SearchQuery: "JavaScript",
					},
					{
						Title:       "Responsive Design",
						Description: "Responsive design is an approach to web design that makes web pages render well on a variety of devices and window or screen sizes.",
						ProTips:     "Learn about media queries and flexible grid layouts.",
						SearchQuery: "Responsive Design in Web Development",
					},
				},
			},
			{
				Title:       "JavaScript Frameworks and Libraries",
				Description: "JavaScript frameworks and libraries are pre-written JavaScript code that helps you build interactive web applications. They provide ready-to-use functions and components that you can use in your code.",
				ProTips:     "Familiarize yourself with the most popular frameworks and libraries.",
				SearchQuery: "JavaScript Frameworks and Libraries",
				Subtopics: []chatGeneratePromptPromptResultSubtopic{
					{
						Title:       "React",
						Description: "React is a JavaScript library for building user interfaces. It is maintained by Facebook and a community of individual developers and companies.",
						ProTips:     "Learn about components and state management.",
						SearchQuery: "React JavaScript Framework",
					},
					{
						Title:       "Vue.js",
						Description: "Vue.js is a progressive JavaScript framework used to build interactive web interfaces. It is designed from the ground up to be incrementally adoptable.",
						ProTips:     "Understand the Vue instance and the Vue CLI.",
						SearchQuery: "Vue.js Framework",
					},
					{
						Title:       "Angular",
						Description: "Angular is a platform and framework for building single-page client applications using HTML and TypeScript. It is maintained by Google.",
						ProTips:     "Learn about components, modules, and services.",
						SearchQuery: "Angular JavaScript Framework",
					},
					{
						Title:       "Svelte",
						Description: "Svelte is a radical new approach to building user interfaces. It shifts the work of rendering from the browser to the compile step, resulting in faster load times and a better user experience.",
						ProTips:     "Understand the Svelte compiler and reactivity.",
						SearchQuery: "Svelte JavaScript Framework",
					},
					{
						Title:       "Node.js",
						Description: "Node.js is an open-source, cross-platform, JavaScript runtime environment that executes JavaScript code outside a web browser. It is used to build scalable network applications.",
						ProTips:     "Learn about the event loop and non-blocking I/O.",
						SearchQuery: "Node.js",
					},
				},
			},
		},
	}

	sb.WriteString(`You are an expert in creating structured learning roadmaps for a learning application. The roadmaps you generate are designed to provide users with a clear and organized path for self-learning, not as a course or detailed content provider. The roadmap will:
1. Include a title and description of the main topic to introduce the subject.
2. Break down the topic into topics and subtopics, each with a title and description to explain the focus of the section.
3. Use a maximum of 2 levels of depth for subtopics. Topics can contain subtopics, but subtopics cannot have further nested levels.
4. Be tailored based on user-provided personalization options:
`)

	for _, userPersonalizationOpt := range promptUserPersonalizationOptions {
		sb.WriteString(" - ")
		sb.WriteString(userPersonalizationOpt)
		sb.WriteString("\n")
	}

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

	sb.WriteString("# A Real Case Example:\n")

	sb.WriteString("### Input:\n")
	sb.WriteString("- Topic: Front End Development\n")
	sb.WriteString("- Daily Time Availability: 1 hour/day\n")
	sb.WriteString("- Total Duration: 1 month\n")
	sb.WriteString("- Skill Level: beginner\n")

	sb.WriteString("\n### Output:\n\n")

	exampleResultJSON, err := json.MarshalIndent(exampleResult, "", "    ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal example format")
		return ""
	}

	sb.Write(exampleResultJSON)

	return sb.String()
}
