package models

type User struct {
	MovieRatings map[int]float32 `json:"userRatings"`
}

type SimilarUser struct {
	UserID     int     `json:"userID"`
	Similarity float64 `json:"similarity"`
}
