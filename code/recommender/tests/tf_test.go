package tests

import (
	"math"
	"recommender/algorithms"
	"testing"
)

func TestTF(t *testing.T) {
	document := "Ut: vulputate auctor leo, (nec) posuere urna tempor nec. Ut amet nec, adipiscing."

	expectedTfMap := map[string]float64{
		"urna":       0.076923,
		"tempor":     0.076923,
		"amet":       0.076923,
		"auctor":     0.076923,
		"vulputate":  0.076923,
		"leo":        0.076923,
		"nec":        0.230769,
		"posuere":    0.076923,
		"adipiscing": 0.076923,
		"ut":         0.153846,
	}

	tfMap := algorithms.TF(document)

	if len(tfMap) != len(expectedTfMap) {
		t.Errorf("TF formula: Expected map length: %d, got: %d", len(expectedTfMap), len(tfMap))
		return
	}
	tolerance := 0.000001
	for token, tf := range tfMap {
		if diff := math.Abs(tf - expectedTfMap[token]); diff > tolerance {
			t.Errorf("TF formula: Expected value for token %s => %f. Got %f", token, expectedTfMap[token], tfMap[token])
		}
	}
}
