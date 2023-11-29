package tests

import (
	"math"
	"recommender/algorithms"
	"testing"
)

func TestIDF(t *testing.T) {
	documents := []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"Pellentesque mattis risus ut sapien varius, nec laoreet lectus condimentum.",
		"Maecenas ut tellus eu magna tristique mattis quis quis dolor.",
		"Donec vel ipsum mattis enim sagittis, eleifend enim vel, varius enim.",
	}

	expectedIdfMap := map[string]float64{
		"lorem":        0.60206,
		"ipsum":        0.30103,
		"dolor":        0.30103,
		"sit":          0.60206,
		"amet":         0.60206,
		"consectetur":  0.60206,
		"adipiscing":   0.60206,
		"elit":         0.60206,
		"pellentesque": 0.60206,
		"mattis":       0.124939,
		"risus":        0.60206,
		"ut":           0.30103,
		"sapien":       0.60206,
		"varius":       0.30103,
		"nec":          0.60206,
		"laoreet":      0.60206,
		"lectus":       0.60206,
		"condimentum":  0.60206,
		"maecenas":     0.60206,
		"tellus":       0.60206,
		"eu":           0.60206,
		"magna":        0.60206,
		"tristique":    0.60206,
		"quis":         0.60206,
		"donec":        0.60206,
		"vel":          0.60206,
		"enim":         0.60206,
		"sagittis":     0.60206,
		"eleifend":     0.60206,
	}

	idfMap := algorithms.IDF(documents)

	if len(idfMap) != len(expectedIdfMap) {
		t.Errorf("IDF formula: Expected map length: %d, got: %d", len(expectedIdfMap), len(idfMap))
		return
	}
	tolerance := 0.000001
	for token, idf := range idfMap {
		if diff := math.Abs(idf - expectedIdfMap[token]); diff > tolerance {
			t.Errorf("IDF formula: Expected value for token %s => %f. Got %f", token, expectedIdfMap[token], idfMap[token])
		}
	}
}
