package algorithms

func JaccardSimilarity(set1, set2 []string) float64 {
	// Convert the sets to maps for faster membership checking.
	set1Map := make(map[string]bool)
	set2Map := make(map[string]bool)

	for _, item := range set1 {
		set1Map[item] = true
	}

	for _, item := range set2 {
		set2Map[item] = true
	}

	// Calculate the intersection size.
	intersection := 0
	for item := range set1Map {
		if set2Map[item] {
			intersection++
		}
	}

	// Calculate the union size.
	union := len(set1Map) + len(set2Map) - intersection

	// Calculate the Jaccard similarity.
	if union == 0 {
		return 0.0 // Avoid division by zero.
	}

	return float64(intersection) / float64(union)
}