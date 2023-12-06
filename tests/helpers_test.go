package tests

import (
	"recommender/helpers"
	"reflect"
	"testing"
)

func TestExtractTokensFromStr(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "Hello, World! This is a test string.",
			expected: []string{"hello", "world", "this", "is", "a", "test", "string"},
		},
		{
			input:    "Some (characters) to remove; another string.",
			expected: []string{"some", "characters", "to", "remove", "another", "string"},
		},
		{
			input:    "Multiple | characters : in ; this \" string \"",
			expected: []string{"multiple", "characters", "in", "this", "string"},
		},
	}

	for _, testCase := range testCases {
		result := helpers.ExtractTokensFromStr(testCase.input)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Errorf("For input '%s', expected tokens: %v, but got: %v", testCase.input, testCase.expected, result)
		}
	}
}
