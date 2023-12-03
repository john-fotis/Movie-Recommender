package main

import (
	"fmt"
	"recommender/algorithms"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"sort"
	"sync"
)

func RecommendBasedOnItem(cfg *config.Config, movies *map[int]model.Movie, userID int) []model.Rating {
	totalRatings := 0
	for _, movie := range *movies {
		totalRatings += len(movie.UserRatings)
	}
	fmt.Printf("Working with %d movie ratings.\n", totalRatings)
	util.StartProfiling("item")
	user := model.User{MovieRatings: make(map[int]float32)}
	// Gather all the user's ratings
	for movieID := range *movies {
		rating, exists := (*movies)[movieID].UserRatings[userID]
		if exists {
			user.MovieRatings[movieID] = rating
		}
	}
	// Find top most similar movies for each movie the user has rated
	similarMoviesMap := make(map[int]map[int]model.SimilarMovie, 0)
	// A movie is recommendable when its similar to at least one movie rated by the selected user
	recommendableMovies := make(map[int]bool, 0)
	for movieID := range user.MovieRatings {
		// Find similar movies only for movies the user liked
		if user.MovieRatings[movieID] >= 4 {
			// Find the top k most similar movies to movieID
			similarMovies := findSimilarMovies(cfg, movieID, movies, k)
			currentSimilarMoviesMap := make(map[int]model.SimilarMovie, len(similarMovies))
			for _, movie := range similarMovies {
				currentSimilarMoviesMap[movie.MovieID] = movie
			}
			for otherMovieID := range currentSimilarMoviesMap {
				// Skip movies the user has already rated
				if _, exists := user.MovieRatings[otherMovieID]; !exists {
					recommendableMovies[otherMovieID] = true
				}
			}
			similarMoviesMap[movieID] = currentSimilarMoviesMap
		}
	}
	// Continue to recommendation part
	ratingForecasts := make([]model.Rating, 0)
	for movieID := range recommendableMovies {
		numerator, denominator := 0.0, 0.0
		for ratedMovieID, similarMovies := range similarMoviesMap {
			if _, exists := similarMovies[movieID]; exists {
				numerator += float64(user.MovieRatings[ratedMovieID]) * float64(similarMovies[movieID].Similarity)
				denominator += float64(similarMovies[movieID].Similarity)
			}
		}
		ratingForecasts = append(ratingForecasts, model.Rating{
			MovieID: movieID,
			Rating:  float32(numerator / denominator),
		})
	}
	// Sort recommended movies by forecasted rating in descending order
	sort.SliceStable(ratingForecasts, func(i, j int) bool {
		return ratingForecasts[i].Rating > ratingForecasts[j].Rating
	})
	if len(ratingForecasts) > cfg.Recommendations {
		ratingForecasts = ratingForecasts[:cfg.Recommendations]
	}
	util.StopProfiling()
	return ratingForecasts
}

func findSimilarMovies(cfg *config.Config, selectedMovieID int, movies *map[int]model.Movie, maxMovies ...int) []model.SimilarMovie {
	moviesToKeep := -1
	if len(maxMovies) > 0 {
		moviesToKeep = maxMovies[0]
	}
	selectedMovie := (*movies)[selectedMovieID]
	// Slice of userIDs who have rated the selected movie
	selectedMovieUsers := make([]int, 0, len(selectedMovie.UserRatings))
	for userID := range selectedMovie.UserRatings {
		selectedMovieUsers = append(selectedMovieUsers, userID)
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	// Divide movies into chunks to split the workload to multiple routines
	movieIDs := make([]int, 0, len(*movies))
	for movieID := range *movies {
		if movieID == selectedMovieID {
			continue
		}
		// Skip movie if it has no ratings
		if len((*movies)[movieID].UserRatings) == 0 {
			continue
		}
		movieIDs = append(movieIDs, movieID)
	}
	if numThreads > len(movieIDs) {
		numThreads = len(movieIDs)
	}
	movieChunks := util.GenerateChunkFromSet(movieIDs, numThreads)
	// The final slice of similar movies from all routines
	similarMovies := make([]model.SimilarMovie, 0, len(*movies))
	for _, movieChunk := range movieChunks {
		wg.Add(1)
		go func(movieIDs []int) {
			defer wg.Done()
			// Slice of similar movies for the current routine
			localSimilarMovies := make([]model.SimilarMovie, 0, len(movieIDs))
			for _, otherMovieID := range movieIDs {
				otherMovie := (*movies)[otherMovieID]
				// Slice of userIDs who have rated the other movie
				otherMovieUsers := make([]int, 0, len(otherMovie.UserRatings))
				for userID := range otherMovie.UserRatings {
					otherMovieUsers = append(otherMovieUsers, userID)
				}
				// Skip otherMovie if it has no common users rating it with selectedMovie
				if len(algorithms.Intersection[int](selectedMovieUsers, otherMovieUsers)) == 0 {
					continue
				}
				// Finally, calculate the similarity using the requested similarity metric
				var similarity float64
				switch cfg.Similarity {
				case "jaccard":
					similarity = algorithms.JaccardSimilarity[int](selectedMovieUsers, otherMovieUsers)
				case "dice":
					similarity = algorithms.DiceSimilarity[int](selectedMovieUsers, otherMovieUsers)
				case "cosine":
					vectorA, vectorB := util.GetUserRatingVectors(selectedMovie, otherMovie, selectedMovieUsers, otherMovieUsers)
					similarity = algorithms.CosineSimilarity[float32](vectorA, vectorB, algorithms.DotProductFloat32)
				case "pearson":
					vectorA, vectorB := util.GetUserRatingVectors(selectedMovie, otherMovie, selectedMovieUsers, otherMovieUsers)
					similarity = algorithms.PearsonSimilarity[float32](vectorA, vectorB)
				}
				localSimilarMovies = append(localSimilarMovies, model.SimilarMovie{
					MovieID: otherMovieID, Similarity: similarity,
				})
			}
			// Merge all local slices of similarMovies while protecting concurrent writing to shared struct
			mu.Lock()
			similarMovies = append(similarMovies, localSimilarMovies...)
			mu.Unlock()
		}(movieChunk)
	}
	// Wait for all routines to finish
	wg.Wait()
	updateMostSimilarMovies(&similarMovies, moviesToKeep)
	return similarMovies
}

func updateMostSimilarMovies(movies *[]model.SimilarMovie, moviesToKeep int) {
	/*
		The following sort ensures that the same set of moviesToKeep movies will be
		returned everytime. However this is based on the arbitrary assumption that
		we sort	movies primarily on their similarity and secondarily by their ID.
		Effectivelly this method tends to promote newer movies over older ones.
		If we just use the similarity factor the last movies in the moviesToKeep-set
		will be ordered randomly, thus affecting slightly the final similar movies map.
	*/
	// sort.SliceStable((*movies), func(i, j int) bool {
	// 	if (*movies)[i].Similarity == (*movies)[j].Similarity {
	// 		return (*movies)[i].MovieID > (*movies)[j].MovieID
	// 	}
	// 	return (*movies)[i].Similarity > (*movies)[j].Similarity
	// })
	// Sort similar movies by similarity in descending order and limit its size to moviesToKeep
	sort.SliceStable((*movies), func(i, j int) bool {
		return (*movies)[i].Similarity > (*movies)[j].Similarity
	})
	if moviesToKeep != -1 && len(*movies) > moviesToKeep {
		(*movies) = (*movies)[:moviesToKeep]
	}
}
