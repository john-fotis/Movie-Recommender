package algorithms

import (
	"fmt"
	"log"
	"math"
	"reflect"
)

var floatType = reflect.TypeOf(float64(0))

// https://en.wikipedia.org/wiki/Pearson_correlation_coefficient#For_a_sample
func PearsonSimilarity[T comparable](vector1 []T, vector2 []T) float64 {
	if len(vector1) != len(vector2) {
		return 0.0
	}

	sumX, sumY, sumXY, sumXsq, sumYsq := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(vector1); i++ {
		x, err := toFloat64(vector1[i])
		if err != nil {
			log.Fatalf("Error: %v", err)
			return float64(math.NaN())
		}
		y, err := toFloat64(vector2[i])
		if err != nil {
			log.Fatalf("Error: %v", err)
			return float64(math.NaN())
		}
		sumX += x
		sumY += y
		sumXY += x * y
		sumXsq += x * x
		sumYsq += y * y
	}

	// Calculate the Pearson correlation coefficient
	numerator := (float64(len(vector1))*sumXY - sumX*sumY)
	denominator := math.Sqrt((float64(len(vector1))*sumXsq - sumX*sumX) * (float64(len(vector1))*sumYsq - sumY*sumY))

	if denominator == 0.0 {
		return 0.0
	}

	return numerator / denominator
}

func toFloat64(value interface{}) (float64, error) {
	v := reflect.ValueOf(value)
	v = reflect.Indirect(v)

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Bool:
		if v.Bool() {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		if !v.Type().ConvertibleTo(floatType) {
			return math.NaN(), fmt.Errorf("Cannot convert %v to float64", v.Type())
		}
		fv := v.Convert(floatType)
		return fv.Float(), nil
	}
}
