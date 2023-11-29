package algorithms

func Union[T comparable](set1 []T, set2 []T) []T {
    totalValues := []T{}
    set1Map := make(map[T]struct{})

    for _, value := range set1 {
        totalValues = append(totalValues, value)
        set1Map[value] = struct{}{}
    }

    for _, value := range set2 {
        if _, exists := set1Map[value]; !exists {
            totalValues = append(totalValues, value)
        }
    }

    return totalValues
}