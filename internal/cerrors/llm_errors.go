package cerrors

import "net/http"

var (
	// ErrLLMProviderUnavailable occurs when the 3rd Party LLM provider is unavailable.
	// This can happen due to various reasons like network issues, provider downtime, etc.
	ErrLLMProviderUnavailable = &curionaError{
		Code:            "LLM_PROVIDER_UNAVAILABLE",
		StatusCode:      http.StatusServiceUnavailable,
		InternalMessage: "LLM Provider Unavailable",
		ExternalMessage: "Oops! Looks like our LLM provider is currently unavailable. Please try again later.",
	}

	// ErrLLMPromptGenerationFailed occurs when constructing a system/user prompt failed.
	// It can be due to various reasons like invalid formatting, marshal/unmarshal errors, etc.
	ErrLLMPromptGenerationFailed = &curionaError{
		Code:            "LLM_PROMPT_GENERATION_FAILED",
		StatusCode:      http.StatusInternalServerError,
		InternalMessage: "Prompt Generation Failed",
		ExternalMessage: "Oops! We encountered an unexpected error while generating the prompt for you. Please try again later.",
	}

	// ErrLLMInvalidData is an error that occurs when the user provides invalid data.
	ErrLLMInvalidData = &curionaError{
		Code:            "LLM_INVALID_DATA",
		StatusCode:      http.StatusBadRequest,
		InternalMessage: "LLM Invalid Data Provided",
		ExternalMessage: "There was an issue with the data you provided. Please check and try again.",
	}

	// ErrLLMMaximumRoadmapGenerationReached occurs when the user has reached the maximum limit for roadmap generations.
	ErrLLMMaximumRoadmapGenerationReached = &curionaError{
		Code:            "LLM_MAXIMUM_ROADMAP_GENERATION_REACHED",
		StatusCode:      http.StatusTooManyRequests,
		InternalMessage: "Maximum Roadmap Generation Limit Reached",
		ExternalMessage: "You have reached the maximum limit for roadmap generations. Either finish a roadmap or delete an existing one to continue.",
	}

	// ErrLLMFlaggedContentDetected occurs when the LLM detects flagged content in the user's input.
	ErrLLMFlaggedContentDetected = &curionaError{
		Code:            "LLM_FLAGGED_CONTENT_DETECTED",
		StatusCode:      http.StatusForbidden,
		InternalMessage: "Flagged Content Detected",
		ExternalMessage: "Your prompt contains content that violates our guidelines. Please modify your prompt and try again.",
	}
)
