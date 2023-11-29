package tests

import (
	"math"
	"recommender/algorithms"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	vector1 := []float32{3.4, 4.7, 1.5, 2.0, 5.0}
	vector2 := []float32{4.0, 5.0, 1.4, 2.5, 4.0}

	result := algorithms.CosineSimilarity(vector1, vector2, algorithms.DotProductFloat32)

	tolerance := 0.000001
	if diff := math.Abs(result - 0.986860); diff > tolerance {
		t.Errorf("Cosine similarity: Expected 0.986860, got %f", result)
		return
	}
}
