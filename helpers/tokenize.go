package helpers

import "strings"

// Removes irrelevant chars from str, makes it lowercase and returns its tokens
func ExtractTokensFromStr(str string) []string {
	irrelevantChars := []string{".", ",", ":", "(", ")", "\"", "'", "|", "!", "?", "#", ";"}
	for _, char := range irrelevantChars {
		str = strings.ReplaceAll(str, char, "")
	}
	return strings.Fields(strings.ToLower(str))
}
