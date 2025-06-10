package casing

import (
	"strings"
	"unicode"
)

func ToKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		switch {
		case r == '_':
			// Replace underscore with hyphen
			result.WriteRune('-')
		case i > 0 && unicode.IsUpper(r):
			// If current letter is uppercase and...
			if unicode.IsLower(rune(s[i-1])) || // previous letter is lowercase
				(i+1 < len(s) && unicode.IsLower(rune(s[i+1]))) { // or next letter is lowercase
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(r))
		default:
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}

func KebabToTitleCase(s string) string {
	var result strings.Builder
	capitalize := true

	for _, r := range s {
		switch {
		case r == '-':
			// Replace hyphen with space
			result.WriteRune(' ')
			capitalize = true
		case capitalize:
			// Capitalize the first letter after a hyphen or at the beginning
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		default:
			// Keep other letters as they are
			result.WriteRune(r)
		}
	}

	return result.String()
}

func SnakeToTitleCase(s string) string {
	var result strings.Builder
	capitalize := true

	for _, r := range s {
		switch {
		case r == '_':
			// Replace underscore with space
			result.WriteRune(' ')
			capitalize = true
		case capitalize:
			// Capitalize the first letter after an underscore or at the beginning
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		default:
			// Keep other letters as they are
			result.WriteRune(r)
		}
	}

	return result.String()
}
