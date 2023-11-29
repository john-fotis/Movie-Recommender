package models

type Movie struct {
	UserRatings map[int]float32 `json:"userRatings"`
}

type SimilarMovie struct {
	MovieID    int     `json:"movieID"`
	Similarity float64 `json:"similarity"`
}
