package tests

import (
	"math"
	"recommender/algorithms"
	"testing"
)

func TestDiceSimilarity(t *testing.T) {
	set1 := []int{1, 2, 3, 7, 10, 20, 100}
	set2 := []int{2, 3, 4, 17, 20}

	result := algorithms.DiceSimilarity[int](set1, set2)

	tolerance := 0.000001
	if diff := math.Abs(result - 0.5); diff > tolerance {
		t.Errorf("Dice similarity: Expected 0.5, got %f", result)
		return
	}
}
