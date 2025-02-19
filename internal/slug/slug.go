package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/unicode/norm"
)

// Make converts a regular string into slug format.
// it does not support multi-language translation and non-latin characters.
func Make(value string) string {
	var buffer strings.Builder

	value = norm.NFD.String(value)
	value = strings.ToLower(value)
	value = runes.Remove(runes.In(unicode.Mn)).String(value)

	spaceFound := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if spaceFound {
				buffer.WriteRune('-')
				spaceFound = false
			}
			buffer.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '_' || r == '-' {
			if !spaceFound {
				spaceFound = true
			}
		}
	}

	slug := buffer.String()
	slug = strings.Trim(slug, "-")
	return slug
}
