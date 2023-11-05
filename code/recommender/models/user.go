package models

type User struct {
	UserRatings []UserRating `json:"userRatings"`
}

type UserRating struct {
	MovieID int     `json:"movieId"`
	Rating  float32 `json:"rating"`
}
