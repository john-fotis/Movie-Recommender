package algorithms

import (
	"recommender/helpers"
)

// Returns a map of {token:TFScore} pairs for every token in the given document.
func TF(document string) map[string]float64 {
	tfMap := make(map[string]float64)
	tokens := helpers.ExtractTokensFromStr(document)
	// Calculate the frequency of each token in the document
	for _, token := range tokens {
		tfMap[token]++
	}
	// Calculate TF by dividing the count of each token by the total number of terms
	for token, occurrences := range tfMap {
		tfMap[token] = occurrences / float64(len(tokens))
	}
	return tfMap
}
