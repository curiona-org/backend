package cerrors

import "net/http"

var (
	ErrInvalidCredentials = &curionaError{
		Code:            "INVALID_CREDENTIALS",
		StatusCode:      http.StatusUnauthorized,
		InternalMessage: "Invalid credentials",
		ExternalMessage: "The credentials you entered are incorrect. Please try again.",
	}

	ErrUnauthorized = &curionaError{
		Code:            "UNAUTHORIZED",
		StatusCode:      http.StatusUnauthorized,
		InternalMessage: "Unauthorized",
		ExternalMessage: "You are not authorized to perform this action.",
	}

	ErrForbidden = &curionaError{
		Code:            "FORBIDDEN",
		StatusCode:      http.StatusForbidden,
		InternalMessage: "Forbidden",
		ExternalMessage: "You do not have permission to perform this action.",
	}

	ErrAccountSuspended = &curionaError{
		Code:            "ACCOUNT_SUSPENDED",
		StatusCode:      http.StatusUnauthorized,
		InternalMessage: "Account is suspended",
		ExternalMessage: "Your account has been suspended. Please contact support for assistance.",
	}

	ErrSignUpDifferentMethod = &curionaError{
		Code:            "SIGNUP_DIFFERENT_METHOD",
		StatusCode:      http.StatusConflict,
		InternalMessage: "User already signed up with a different method",
		ExternalMessage: "An account with this email already exists, but it uses a different sign-in method. Please try signing in with another method.",
	}

	ErrSessionExpired = &curionaError{
		Code:            "SESSION_EXPIRED",
		StatusCode:      http.StatusUnauthorized,
		InternalMessage: "Session expired",
		ExternalMessage: "Your session has expired. Please sign in again.",
	}

	ErrSessionIsBlocked = &curionaError{
		Code:            "SESSION_BLOCKED",
		StatusCode:      http.StatusUnauthorized,
		InternalMessage: "Session is blocked",
		ExternalMessage: "Your session has been blocked. Please sign in again.",
	}
)
