package algorithms

// DiceSimilarity calculates the Dice similarity coefficient between two sets.
func DiceSimilarity(set1, set2 []string) float64 {
	// Create maps to represent the sets for efficient intersection and size calculations.
	set1Map := make(map[string]struct{})
	set2Map := make(map[string]struct{})

	// Populate set1Map and set2Map with elements from set1 and set2.
	for _, item := range set1 {
		set1Map[item] = struct{}{}
	}
	for _, item := range set2 {
		set2Map[item] = struct{}{}
	}

	// Calculate the intersection size.
	intersectionSize := 0
	for item := range set1Map {
		if _, exists := set2Map[item]; exists {
			intersectionSize++
		}
	}

	// Calculate the Dice similarity coefficient.
	if len(set1Map) == 0 && len(set2Map) == 0 {
		return 1.0 // Both sets are empty, so they are considered equal.
	} else {
		return (2.0 * float64(intersectionSize)) / (float64(len(set1Map)) + float64(len(set2Map)))
	}
}