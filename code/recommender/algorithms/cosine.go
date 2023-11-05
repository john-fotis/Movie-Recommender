package algorithms

import "math"

// CosineSimilarity calculates the cosine similarity between two vectors.
func CosineSimilarity(vector1, vector2 []float64) float64 {
	// Check if the vectors are of the same length.
	if len(vector1) != len(vector2) {
		return 0.0 // Cosine similarity is not defined for vectors of different lengths.
	}

	// Calculate the dot product and magnitudes of the vectors.
	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	for i := 0; i < len(vector1); i++ {
		dotProduct += vector1[i] * vector2[i]
		magnitude1 += vector1[i] * vector1[i]
		magnitude2 += vector2[i] * vector2[i]
	}

	// Calculate the cosine similarity.
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0 // Cosine similarity is not defined for zero vectors.
	}

	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}
