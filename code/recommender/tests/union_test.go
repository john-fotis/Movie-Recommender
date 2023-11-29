package tests

import (
	"recommender/algorithms"
	"sort"
	"testing"
)

func TestUnion(t *testing.T) {
	set1 := []int{1, 2, 3, 7, 10, 20, 100}
	set2 := []int{2, 3, 4, 17, 20}

	result := algorithms.Union(set1, set2)
	expectedResult := []int{1, 2, 3, 4, 7, 10, 17, 20, 100}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	if len(result) != len(expectedResult) {
		t.Errorf("Union formula: Expected %v, got %v", expectedResult, result)
		return
	}

	for index, item := range result {
		if item != expectedResult[index] {
			t.Errorf("Union formula: expected %v, got %v", expectedResult, result)
		}
	}
}
