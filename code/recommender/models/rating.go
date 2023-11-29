package models

type Rating struct {
	// This struct stores a rating for a specific movie independently of specific user ID(s)
	// The user ID is implicit and meant to be used as a value in a {userID:rating} pair
	MovieID int     `json:"movieId"`
	Rating  float32 `json:"rating"`
}
