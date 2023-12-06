package algorithms

import (
	"math"
	"recommender/helpers"
)

// Returns a map of {token:IDFScore} pairs for every token in every document
func IDF(documents []string) map[string]float64 {
	// Characters to remove from the documents
	idfMap := make(map[string]float64)
	// Calculate the occurence of each token in each document
	for _, document := range documents {
		processedTokens := make(map[string]bool)
		for _, token := range helpers.ExtractTokensFromStr(document) {
			if processedTokens[token] {
				continue
			}
			idfMap[token]++
			processedTokens[token] = true
		}
	}
	// Calculate IDF for each token based on its occurences in the documents
	for token, occurences := range idfMap {
		idfMap[token] = math.Log10(float64(len(documents)) / float64(occurences))
	}
	return idfMap
}
