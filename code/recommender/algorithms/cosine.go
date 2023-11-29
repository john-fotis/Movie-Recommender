package algorithms

import (
	"math"
	"reflect"
)

type DotProductFunc[T comparable] func(T, T) float64

// https://en.wikipedia.org/wiki/Cosine_similarity
func CosineSimilarity[T comparable](vector1 []T, vector2 []T, dotProduct DotProductFunc[T]) float64 {
	if len(vector1) != len(vector2) {
		return 0.0
	}

	dotProductSum := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0
	for i := 0; i < len(vector1); i++ {
		dotProductSum += dotProduct(vector1[i], vector2[i])
		magnitude1 += dotProduct(vector1[i], vector1[i])
		magnitude2 += dotProduct(vector2[i], vector2[i])
	}

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	return dotProductSum / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}

func DotProductInt(a int, b int) float64 {
	return float64(a) * float64(b)
}

func DotProductFloat32(a float32, b float32) float64 {
	return float64(a) * float64(b)
}

func DotProductFloat64(a float64, b float64) float64 {
	return float64(a) * float64(b)
}

func DotProductBool(a bool, b bool) float64 {
	if a && b {
		return 1.0
	}
	return 0.0
}

func IsNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
