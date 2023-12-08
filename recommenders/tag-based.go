package recommenders

import (
	"fmt"
	"recommender/algorithms"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"sort"
	"sync"
)

var cachedTagStrings sync.Map

func RecommendBasedOnTag(cfg *config.Config, movieTags *map[int]model.MovieTags) []model.SimilarMovie {
	totalTags := 0
	for movieID := range *movieTags {
		for _, userTags := range (*movieTags)[movieID].UserTags {
			totalTags += len(userTags.Tags)
		}
	}
	fmt.Printf("Working with %d movie tags.\n", totalTags)
	util.StartProfiling("tag")
	selectedMovieTags := make([]string, 0)
	for _, userTags := range (*movieTags)[cfg.Input].UserTags {
		selectedMovieTags = append(selectedMovieTags, userTags.Tags...)
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	// Divide movies into chunks to split the workload to multiple routines
	movieIDs := make([]int, 0, len(*movieTags))
	for movieID := range *movieTags {
		if movieID == cfg.Input {
			continue
		}
		// Movies with not at least 1 common tag will always have 0 similarity
		if len(algorithms.Intersection[string](
			getAllMovieTagStrings(cfg.Input, (*movieTags)[cfg.Input]),
			getAllMovieTagStrings(movieID, (*movieTags)[movieID]),
		)) == 0 {
			continue
		}
		movieIDs = append(movieIDs, movieID)
	}
	if cfg.NumThreads > len(movieIDs) {
		cfg.NumThreads = len(movieIDs)
	}
	movieChunks := util.GenerateChunkFromSet(movieIDs, cfg.NumThreads)
	// The final slice of similar movies from all routines
	similarMovies := make([]model.SimilarMovie, 0, len(*movieTags))
	for _, movieChunk := range movieChunks {
		wg.Add(1)
		go func(movieIDs []int) {
			defer wg.Done()
			// Slice of similar movies for the current routine
			localSimilarMovies := make([]model.SimilarMovie, 0, len(movieIDs))
			for _, otherMovieID := range movieIDs {
				otherMovieTags := make([]string, 0)
				otherMovieUsersThatTagged := (*movieTags)[otherMovieID].UserTags
				for _, userTags := range otherMovieUsersThatTagged {
					otherMovieTags = append(otherMovieTags, userTags.Tags...)
				}
				// Skip movie if it has no tags
				if len(otherMovieTags) == 0 {
					continue
				}
				// Finally, calculate the similarity using the requested similarity metric
				var similarity float64
				switch cfg.Similarity {
				case "jaccard":
					similarity = algorithms.JaccardSimilarity[string](selectedMovieTags, otherMovieTags)
				case "dice":
					similarity = algorithms.DiceSimilarity[string](selectedMovieTags, otherMovieTags)
				case "cosine":
					vectorA, vectorB := util.GetTagOccurenceVectors(
						countTagOccurrences((*movieTags)[cfg.Input]),
						countTagOccurrences((*movieTags)[otherMovieID]),
					)
					similarity = algorithms.CosineSimilarity[int](vectorA, vectorB, algorithms.DotProductInt)
				case "pearson":
					vectorA, vectorB := util.GetTagOccurenceVectors(
						countTagOccurrences((*movieTags)[cfg.Input]),
						countTagOccurrences((*movieTags)[otherMovieID]),
					)
					similarity = (algorithms.PearsonSimilarity[int](vectorA, vectorB) + 1) / 2
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

func getAllMovieTagStrings(movieID int, movieTags model.MovieTags) []string {
	// Generate a unique key for the movie to use in the cache
	cacheKey := fmt.Sprintf("%d", movieID)
	// Check if the tag strings are already cached
	if cachedTags, exist := cachedTagStrings.Load(cacheKey); exist {
		if cachedTagsSlice, exist := cachedTags.([]string); exist {
			return cachedTagsSlice
		}
	}
	// If tags aren't cached, get them and cache them for future use
	movieTagStrings := gatherMovieTagStrings(movieTags)
	cachedTagStrings.Store(cacheKey, movieTagStrings)
	return movieTagStrings
}

// Function to gather the tag strings for a movie
func gatherMovieTagStrings(movie model.MovieTags) []string {
	tagStrings := make([]string, 0)
	// Retrieving tags from countTagOccurrences ensures each tag is appended at most once
	for tag := range countTagOccurrences(movie) {
		tagStrings = append(tagStrings, tag)
	}
	return tagStrings
}

func countTagOccurrences(movieTags model.MovieTags) map[string]int {
	tagCounts := make(map[string]int)
	for _, userTags := range movieTags.UserTags {
		for _, tag := range userTags.Tags {
			tagCounts[tag]++
		}
	}
	return tagCounts
}
