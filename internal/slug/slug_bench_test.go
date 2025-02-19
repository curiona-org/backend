package slug_test

import (
	"testing"

	"github.com/roadmap-thesis/backend/internal/slug"
)

func BenchmarkSanitize(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"LettersAndDigits", "HelloWorld123"},
		{"Spaces", "Hello World"},
		{"SpecialCharacters", "Hello@World!"},
		{"MixedCharacters", "Hello World! 123"},
		{"EmptyString", ""},
		{"LeadingAndTrailingSpaces", "  Hello World  "},
		{"MultipleSpaces", "Hello   World"},
		{"Underscores", "Hello_World"},
		{"Hyphens", "Hello-World"},
		{"MixedSeparators", "Hello_World-123"},
		{"AccentedCharacters", "Héllo Wörld"},
		{"NumbersOnly", "1234567890"},
		{"UppercaseLetters", "HELLO WORLD"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for range b.N {
				slug.Make(tc.input)
			}
		})
	}
}
