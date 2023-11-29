package main

import (
	"fmt"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"sort"
)

func RecommendHybrid(cfg *config.Config, selectedUserMovies map[int]float32, titles *map[int]model.MovieTitle, movies *map[int]model.Movie, tags *map[int]model.MovieTags) []model.SimilarMovie {
	util.StartProfiling("hybrid")
	finalSimilarMovies := make([]model.SimilarMovie, 0)
	knownMovieStats := make(map[int]struct {
		Sum   float64
		Count int
	})
	// Combine title & tag collaborative filtering
	for ratedMovieID, rating := range selectedUserMovies {
		if rating < 4 {
			continue
		}
		titleCgf := config.Config{Recommendations: len(*titles), Similarity: cfg.Similarity, Input: ratedMovieID}
		similarMoviesByTitle := RecommendBasedOnTitle(&titleCgf, titles)
		tagCfg := config.Config{Recommendations: len(*tags), Similarity: cfg.Similarity, Input: ratedMovieID}
		similarMoviesByTag := RecommendBasedOnTag(&tagCfg, tags)
		for _, movieByTitle := range similarMoviesByTitle {
			for _, movieByTag := range similarMoviesByTag {
				if movieByTitle.MovieID == movieByTag.MovieID {
					// Combine title and tag similarities for each movie using 67% - 33% weight respectivelly
					currentSimilarity := 0.67*movieByTitle.Similarity + 0.33*movieByTag.Similarity
					if existingStats, exists := knownMovieStats[movieByTitle.MovieID]; exists {
						// If the movie has already some stats, this means it came out similar to at least
						// one more movie the user has rated. Append data to calc average stats at the end
						existingStats.Sum += currentSimilarity
						existingStats.Count++
						knownMovieStats[movieByTitle.MovieID] = existingStats
					} else {
						knownMovieStats[movieByTitle.MovieID] = struct {
							Sum   float64
							Count int
						}{
							Sum:   currentSimilarity,
							Count: 1,
						}
					}
				}
			}
		}
	}
	// Calculate average similarities for all movies similar to any of the rated movies
	for movieID, movieStats := range knownMovieStats {
		avgSimilarity := movieStats.Sum / float64(movieStats.Count)
		finalSimilarMovies = append(finalSimilarMovies, model.SimilarMovie{
			MovieID:    movieID,
			Similarity: avgSimilarity,
		})
	}
	sort.SliceStable(finalSimilarMovies, func(i, j int) bool {
		return finalSimilarMovies[i].Similarity > finalSimilarMovies[j].Similarity
	})
	if len(finalSimilarMovies) > cfg.Recommendations {
		finalSimilarMovies = finalSimilarMovies[:cfg.Recommendations]
	}
	for i, similarMovie := range finalSimilarMovies {
		if similarMovie.Similarity > 0.1 {
			fmt.Printf("%d: %s => %.2f\n", i+1, (*titles)[similarMovie.MovieID].Title, similarMovie.Similarity)
		}
	}
	util.StopProfiling()
	return finalSimilarMovies
}

func medianRatedMoviesPerUser(users *map[int]model.User) int {
	userRatedMovies := make([]int, 0, len(*users))
	// Store the number of rated movies for each user
	for _, user := range *users {
		userRatedMovies = append(userRatedMovies, len(user.MovieRatings))
	}
	sort.SliceStable(userRatedMovies, func(i, j int) bool {
		return userRatedMovies[i] > userRatedMovies[j]
	})
	// Pick the median number of rated movies from the sorted slice
	return userRatedMovies[len(userRatedMovies)/2]
}
