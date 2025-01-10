package utils

import (
	"strings"
	"unicode"
)

func GetStringIntials(input string) string {
	words := strings.Fields(input)

	var initials string

	for i := 0; i < len(words); i++ {
		initial := string(unicode.ToUpper(rune(words[i][0])))

		initials += initial
	}

	return initials
}
