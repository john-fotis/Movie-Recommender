package models

type MovieTags struct {
	UserTags map[int]UserTags `json:"userTags"`
}

type UserTags struct {
	Tags []string `json:"tags"`
}

type TagOccurrence struct {
	Tag         string `json:"tag"`
	Occurrences int    `json:"occurences"`
}
