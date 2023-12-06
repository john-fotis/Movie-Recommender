package algorithms

func Intersection[T comparable](set1 []T, set2 []T) []T {
	commonValues := make([]T, 0)
	set1Map := make(map[T]struct{})

	for _, value := range set1 {
		set1Map[value] = struct{}{}
	}

	for _, value := range set2 {
		if _, exists := set1Map[value]; exists {
			commonValues = append(commonValues, value)
		}
	}

	return commonValues
}