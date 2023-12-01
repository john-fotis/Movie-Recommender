package main

import (
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"sort"
)

func RecommendHybrid(cfg *config.Config, titles *map[int]model.MovieTitle, movies *map[int]model.Movie, tags *map[int]model.MovieTags) []model.SimilarMovie {
	finalSimilarMovies := make([]model.SimilarMovie, 0)
	knownMovieStats := make(map[int]struct {
		Sum   float64
		Count int
	})
	// Combine item-item, title & tag collaborative filtering and sort all
	// result-slices for more efficient calulcations of average similarity.
	movieCfg := config.Config{Similarity: cfg.Similarity}
	similarMovies := findSimilarMovies(&movieCfg, cfg.Input, movies)
	sort.SliceStable(similarMovies, func(i, j int) bool {
		return similarMovies[i].MovieID < similarMovies[j].MovieID
	})
	titleCgf := config.Config{Recommendations: len(*titles), Similarity: cfg.Similarity, Input: cfg.Input}
	similarMoviesByTitle := RecommendBasedOnTitle(&titleCgf, titles)
	sort.SliceStable(similarMoviesByTitle, func(i, j int) bool {
		return similarMoviesByTitle[i].MovieID < similarMoviesByTitle[j].MovieID
	})
	tagCfg := config.Config{Recommendations: len(*tags), Similarity: cfg.Similarity, Input: cfg.Input}
	similarMoviesByTag := RecommendBasedOnTag(&tagCfg, tags)
	sort.SliceStable(similarMoviesByTag, func(i, j int) bool {
		return similarMoviesByTag[i].MovieID < similarMoviesByTag[j].MovieID
	})
	// Merge the similarities of all 3 algorithms
	idxMovie, idxTitle, idxTag := 0, 0, 0
	for idxMovie < len(similarMovies) && idxTitle < len(similarMoviesByTitle) && idxTag < len(similarMoviesByTag) {
		similarMovie := similarMovies[idxMovie]
		similarMovieByTitle := similarMoviesByTitle[idxTitle]
		similarMovieByTag := similarMoviesByTag[idxTag]
		if similarMovie.MovieID == similarMovieByTitle.MovieID && similarMovie.MovieID == similarMovieByTag.MovieID {
			// Combine the 3 similarity scores using 10%-50%-40% weights
			currentSimilarity := 0.1*similarMovie.Similarity*0.5*similarMovieByTitle.Similarity + 0.4*similarMovieByTag.Similarity
			if existingStats, exists := knownMovieStats[similarMovie.MovieID]; exists {
				// If the movie has already some stats, this means it came out similar to at least
				// one more movie the user has rated. Append data to calc average stats at the end.
				existingStats.Sum += currentSimilarity
				existingStats.Count++
				knownMovieStats[similarMovie.MovieID] = existingStats
			} else {
				knownMovieStats[similarMovie.MovieID] = struct {
					Sum   float64
					Count int
				}{
					Sum:   currentSimilarity,
					Count: 1,
				}
			}
			// Move to the next item in all slices
			idxMovie++
			idxTitle++
			idxTag++
		} else {
			// Move indices of each slice based on MovieID
			if similarMovie.MovieID < similarMovieByTitle.MovieID {
				idxMovie++
			}
			if similarMovieByTitle.MovieID < similarMovieByTag.MovieID {
				idxTitle++
			}
			if similarMovieByTag.MovieID < similarMovie.MovieID {
				idxTag++
			}
		}
	}
	// Calculate average similarities for all movies similar to the requested movie
	for movieID, movieStats := range knownMovieStats {
		finalSimilarMovies = append(finalSimilarMovies, model.SimilarMovie{
			MovieID:    movieID,
			Similarity: movieStats.Sum / float64(movieStats.Count),
		})
	}
	sort.SliceStable(finalSimilarMovies, func(i, j int) bool {
		return finalSimilarMovies[i].Similarity > finalSimilarMovies[j].Similarity
	})
	if len(finalSimilarMovies) > cfg.Recommendations {
		finalSimilarMovies = finalSimilarMovies[:cfg.Recommendations]
	}
	util.StopProfiling()
	return finalSimilarMovies
}
