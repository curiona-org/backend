package cerrors

import "net/http"

var (
	// ErrLLMProviderUnavailable occurs when the 3rd Party LLM provider is unavailable.
	// This can happen due to various reasons like network issues, provider downtime, etc.
	ErrLLMProviderUnavailable = &AppError{
		ErrorCode:       http.StatusServiceUnavailable,
		InternalMessage: "LLM Provider Unavailable",
		ExternalMessage: "Oops! Looks like our LLM provider is currently unavailable. Please try again later.",
	}

	// ErrPromptGenerationFailed occurs when constructing a system/user prompt failed.
	// It can be due to various reasons like invalid formatting, marshal/unmarshal errors, etc.
	ErrPromptGenerationFailed = &AppError{
		ErrorCode:       http.StatusInternalServerError,
		InternalMessage: "Prompt Generation Failed",
		ExternalMessage: "Oops! We encountered an unexpected error while generating the prompt for you. Please try again later.",
	}

	// ErrLLMInvalidData is an error that occurs when the user provides invalid data.
	ErrLLMInvalidData = &AppError{
		ErrorCode:       http.StatusBadRequest,
		InternalMessage: "LLM Invalid Data Provided",
		ExternalMessage: "Incorrect data provided, please check and try again.",
	}
)
