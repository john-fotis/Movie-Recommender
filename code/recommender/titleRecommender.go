package main

import (
	"recommender/algorithms"
	"recommender/config"
	"recommender/helpers"
	model "recommender/models"
	util "recommender/utils"
	"sort"
	"sync"
)

func RecommendBasedOnTitle(cfg *config.Config, movieTitles *map[int]model.MovieTitle) []model.SimilarMovie {
	util.StartProfiling("title")
	idfMap := make(map[string]float64, 0)
	selectedMovieTFMap := make(map[string]float64, 0)
	selectedMovieTitleTokens := make([]string, 0)
	titleSlice := make([]string, 0, len(*movieTitles))
	for _, movie := range *movieTitles {
		titleSlice = append(titleSlice, movie.Title)
	}
	// Calculate IDF for all movie titles to be used for Cosine or Pearson
	idfMap = algorithms.IDF(titleSlice)
	// Calculate TF vector for the selected movie to be used for Cosine or Pearson
	selectedMovieTFMap = algorithms.TF((*movieTitles)[cfg.Input].Title)
	// Extract selected movie title tokens to be used for Jaccard or Dice
	selectedMovieTitleTokens = helpers.ExtractTokensFromStr((*movieTitles)[cfg.Input].Title)
	var mu sync.Mutex
	var wg sync.WaitGroup
	// Divide movies into chunks to split the workload to multiple routines
	movieIDs := make([]int, 0, len(*movieTitles))
	for movieID := range *movieTitles {
		if movieID == cfg.Input {
			continue
		}
		movieTitleTokens := helpers.ExtractTokensFromStr((*movieTitles)[movieID].Title)
		if len(algorithms.Intersection[string](selectedMovieTitleTokens, movieTitleTokens)) == 0 {
			continue
		}
		movieIDs = append(movieIDs, movieID)
	}
	if numThreads > len(movieIDs) {
		numThreads = len(movieIDs)
	}
	movieChunks := util.GenerateChunkFromSet(movieIDs, numThreads)
	// The final slice of similar movier from all routines
	similarMovies := make([]model.SimilarMovie, 0, len(*movieTitles))
	for _, movieChunk := range movieChunks {
		wg.Add(1)
		go func(movieIDs []int) {
			defer wg.Done()
			// Slice of similar movies for the current routine
			localSimilarMovies := make([]model.SimilarMovie, 0, len(movieIDs))
			for _, otherMovieID := range movieIDs {
				var similarity float64
				switch cfg.Similarity {
				case "jaccard":
					otherMovieTitleTokens := helpers.ExtractTokensFromStr((*movieTitles)[otherMovieID].Title)
					similarity = algorithms.JaccardSimilarity[string](selectedMovieTitleTokens, otherMovieTitleTokens)
				case "dice":
					otherMovieTitleTokens := helpers.ExtractTokensFromStr((*movieTitles)[otherMovieID].Title)
					similarity = algorithms.DiceSimilarity[string](selectedMovieTitleTokens, otherMovieTitleTokens)
				case "cosine":
					othetMovieTFMap := algorithms.TF((*movieTitles)[otherMovieID].Title)
					vectorA, vectorB := util.GetTfIdfVectors(idfMap, selectedMovieTFMap, othetMovieTFMap)
					similarity = algorithms.CosineSimilarity[float64](vectorA, vectorB, algorithms.DotProductFloat64)
				case "pearson":
					othetMovieTFMap := algorithms.TF((*movieTitles)[otherMovieID].Title)
					vectorA, vectorB := util.GetTfIdfVectors(idfMap, selectedMovieTFMap, othetMovieTFMap)
					similarity = algorithms.PearsonSimilarity[float64](vectorA, vectorB)
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
	// Sort recommended movies by similarity in descending order
	sort.SliceStable(similarMovies, func(i, j int) bool {
		return similarMovies[i].Similarity > similarMovies[j].Similarity
	})
	if len(similarMovies) > cfg.Recommendations {
		similarMovies = similarMovies[:cfg.Recommendations]
	}
	util.StopProfiling()
	return similarMovies
}
