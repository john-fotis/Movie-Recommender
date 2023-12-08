package recommenders

import (
	"fmt"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"sort"
)

func RecommendHybrid(cfg *config.Config, titles *map[int]model.MovieTitle, movies *map[int]model.Movie, tags *map[int]model.MovieTags) []model.SimilarMovie {
	totalRatings := 0
	for _, movie := range *movies {
		totalRatings += len(movie.UserRatings)
	}
	fmt.Printf("Working with %d movie ratings.\n", totalRatings)
	util.StartProfiling("hybrid")
	finalSimilarMovies := make([]model.SimilarMovie, 0)
	// Combine tag, title & item-item collaborative filtering and sort all
	// result-slices for more efficient calulcations of average similarity.
	tagCfg := config.Config{Recommendations: len(*tags), Similarity: cfg.Similarity, Input: cfg.Input}
	similarMoviesByTag := RecommendBasedOnTag(&tagCfg, tags)
	sort.SliceStable(similarMoviesByTag, func(i, j int) bool {
		return similarMoviesByTag[i].MovieID < similarMoviesByTag[j].MovieID
	})
	// Create a subset of movie titles, only keeping the movieIDs that are recommendable by Tag-based correlation
	recommendableTitles := map[int]model.MovieTitle{cfg.Input: (*titles)[cfg.Input]}
	for _, movie := range similarMoviesByTag {
		recommendableTitles[movie.MovieID] = ((*titles)[movie.MovieID])
	}
	titleCgf := config.Config{Recommendations: len(*titles), Similarity: cfg.Similarity, Input: cfg.Input}
	similarMoviesByTitle := RecommendBasedOnTitle(&titleCgf, &recommendableTitles)
	sort.SliceStable(similarMoviesByTitle, func(i, j int) bool {
		return similarMoviesByTitle[i].MovieID < similarMoviesByTitle[j].MovieID
	})
	// Create a subset of movies, only keeping the movieIDs that are recommendable by Title-based correlation
	recommendableMovies := map[int]model.Movie{cfg.Input: (*movies)[cfg.Input]}
	for _, movie := range similarMoviesByTitle {
		recommendableMovies[movie.MovieID] = ((*movies)[movie.MovieID])
	}
	movieCfg := config.Config{Similarity: cfg.Similarity, NumThreads: cfg.NumThreads}
	similarMovies := findSimilarMovies(&movieCfg, cfg.Input, &recommendableMovies)
	sort.SliceStable(similarMovies, func(i, j int) bool {
		return similarMovies[i].MovieID < similarMovies[j].MovieID
	})
	// Merge the similarities of all 3 algorithms
	idxMovie, idxTitle, idxTag := 0, 0, 0
	for idxMovie < len(similarMovies) && idxTitle < len(similarMoviesByTitle) && idxTag < len(similarMoviesByTag) {
		similarMovie := similarMovies[idxMovie]
		similarMovieByTitle := similarMoviesByTitle[idxTitle]
		similarMovieByTag := similarMoviesByTag[idxTag]
		if similarMovie.MovieID == similarMovieByTitle.MovieID && similarMovie.MovieID == similarMovieByTag.MovieID {
			// Combine the 3 similarity scores using 20%-40%-40% weights so that the upper limit of similarity remains 1.0
			similarity := 0.2*similarMovie.Similarity + 0.4*similarMovieByTitle.Similarity + 0.4*similarMovieByTag.Similarity
			finalSimilarMovies = append(finalSimilarMovies, model.SimilarMovie{
				MovieID:    similarMovie.MovieID,
				Similarity: similarity,
			})
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
	sort.SliceStable(finalSimilarMovies, func(i, j int) bool {
		return finalSimilarMovies[i].Similarity > finalSimilarMovies[j].Similarity
	})
	if len(finalSimilarMovies) > cfg.Recommendations {
		finalSimilarMovies = finalSimilarMovies[:cfg.Recommendations]
	}
	util.StopProfiling()
	return finalSimilarMovies
}
