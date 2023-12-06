package algorithms

// https://en.wikipedia.org/wiki/Jaccard_index
func JaccardSimilarity[T comparable](set1 []T, set2 []T) float64 {
	intersection := Intersection[T](set1, set2)
	union := Union[T](set1, set2)
	if len(union) == 0.0 {
		return 0.0
	}
	return float64(len(intersection)) / float64(len(union))
}
