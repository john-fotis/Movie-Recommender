package models

type GenomeScore struct {
	MovieID   int     `json:"movieId"`
	TagID     int     `json:"tagId"`
	Relevance float32 `json:"relevance"`
}
