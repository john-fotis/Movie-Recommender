package algorithms

// https://en.wikipedia.org/wiki/S%C3%B8rensen%E2%80%93Dice_coefficient
func DiceSimilarity[T comparable](set1 []T, set2 []T) float64 {
	X := len(set1)
	Y := len(set2)
	intersection := Intersection[T](set1, set2)
	if X+Y == 0 {
		return 0.0
	}
	return float64(2*len(intersection)) / float64(X+Y)
}
