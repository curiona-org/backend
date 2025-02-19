package slug_test

import (
	"testing"

	"github.com/roadmap-thesis/backend/internal/slug"
	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"LettersAndDigits", "HelloWorld123", "helloworld123"},
		{"Spaces", "Hello World", "hello-world"},
		{"SpecialCharacters", "Hello@World!", "helloworld"},
		{"MixedCharacters", "Hello World! 123", "hello-world-123"},
		{"EmptyString", "", ""},
		{"LeadingAndTrailingSpaces", "  Hello World  ", "hello-world"},
		{"MultipleSpaces", "Hello   World", "hello-world"},
		{"Underscores", "Hello_World", "hello-world"},
		{"Hyphens", "Hello-World", "hello-world"},
		{"MixedSeparators", "Hello_World-123", "hello-world-123"},
		{"AccentedCharacters", "Héllo Wörld", "hello-world"},
		{"NumbersOnly", "1234567890", "1234567890"},
		{"UppercaseLetters", "HELLO WORLD", "hello-world"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := slug.Make(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
