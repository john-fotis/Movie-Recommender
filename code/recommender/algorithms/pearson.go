package algorithms

import "math"

// PearsonSimilarity calculates the Pearson correlation coefficient between two vectors.
func PearsonSimilarity(vector1, vector2 []float64) float64 {
	// Check if the vectors are of the same length.
	if len(vector1) != len(vector2) {
		return 0.0 // Pearson correlation is not defined for vectors of different lengths.
	}

	n := len(vector1)

	// Calculate the means of both vectors.
	mean1 := mean(vector1)
	mean2 := mean(vector2)

	// Calculate the Pearson correlation coefficient.
	numerator := 0.0
	denominator1 := 0.0
	denominator2 := 0.0

	for i := 0; i < n; i++ {
		numerator += (vector1[i] - mean1) * (vector2[i] - mean2)
		denominator1 += math.Pow(vector1[i]-mean1, 2)
		denominator2 += math.Pow(vector2[i]-mean2, 2)
	}

	denominator := math.Sqrt(denominator1) * math.Sqrt(denominator2)

	if denominator == 0 {
		return 0.0 // Pearson correlation is not defined for zero variance.
	}

	return numerator / denominator
}

func mean(vector []float64) float64 {
	if len(vector) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range vector {
		sum += value
	}
	return sum / float64(len(vector))
}
