package tests

import (
	"math"
	"recommender/algorithms"
	"testing"
)

func TestPearsonSimilarity(t *testing.T) {
	vector1 := []float32{3.4, 4.7, 1.5, 2.0, 5.0}
	vector2 := []float32{4.0, 5.0, 1.4, 2.5, 4.0}

	result := algorithms.PearsonSimilarity(vector1, vector2)

	tolerance := 0.000001
	if diff := math.Abs(result - 0.909470); diff > tolerance {
		t.Errorf("Pearson similarity: Expected 0.909470, got %f", result)
		return
	}
}
