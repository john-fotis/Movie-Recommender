package models

type Link struct {
	MovieID int    `json:"movieId"`
	IMDbID  string `json:"imdbId"`
	TMDBID  int    `json:"tmdbId"`
}
