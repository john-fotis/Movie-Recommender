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

func RecommendBasedOnUser(cfg *config.Config, users *map[int]model.User, movieTitles *map[int]model.MovieTitle) []model.Rating {
	totalRatings := 0
	for _, user := range *users {
		totalRatings += len(user.MovieRatings)
	}
	fmt.Printf("Working with %d user ratings.\n", totalRatings)
	util.StartProfiling("user")
	selectedUser := (*users)[cfg.Input]
	similarUsers := findSimilarUsers(cfg, users)
	ratingForecasts := make([]model.Rating, 0)
	for movieID := range *movieTitles {
		// Skip movies the user has already rated
		if _, exists := selectedUser.MovieRatings[movieID]; !exists {
			numerator, denominator := float64(0), float64(0)
			for _, similarUser := range similarUsers {
				// Only consider (similar) users who have rated this movie
				if rating, exists := (*users)[similarUser.UserID].MovieRatings[movieID]; exists {
					numerator += float64(rating) * float64(similarUser.Similarity)
					denominator += float64(similarUser.Similarity)
				}
			}
			if denominator != 0 {
				// At least one (similar) user must have rated the movie in order to forecast
				ratingForecasts = append(ratingForecasts, model.Rating{
					MovieID: movieID, Rating: float32(numerator / denominator),
				})
			}
		}
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

func findSimilarUsers(cfg *config.Config, users *map[int]model.User) []model.SimilarUser {
	selectedUser := (*users)[cfg.Input]
	// Slice of movieIDs the selecter user has rated
	selectedUserMovies := make([]int, 0, len(selectedUser.MovieRatings))
	for movieID := range selectedUser.MovieRatings {
		selectedUserMovies = append(selectedUserMovies, movieID)
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	// Divide users into chunks to split the workload to multiple routines
	userIDs := make([]int, 0, len(*users))
	for userID := range *users {
		if userID == cfg.Input {
			continue
		}
		userIDs = append(userIDs, userID)
	}
	if cfg.NumThreads > len(userIDs) {
		cfg.NumThreads = len(userIDs)
	}
	userChunks := util.GenerateChunkFromSet(userIDs, cfg.NumThreads)
	// The final slice of similar users from all routines
	similarUsers := make([]model.SimilarUser, 0, len(*users))
	for _, userChunk := range userChunks {
		wg.Add(1)
		go func(userIDs []int) {
			defer wg.Done()
			// Slice of similar users for the current routine
			localSimilarUsers := make([]model.SimilarUser, 0, len(userIDs))
			// Calculate the similarity to $selectedUser for every other user in $userChunk
			for _, otherUserID := range userIDs {
				user := (*users)[otherUserID]
				userMovies := make([]int, 0, len(user.MovieRatings))
				// Store all movies rated by the current user
				for movieID := range user.MovieRatings {
					userMovies = append(userMovies, movieID)
				}
				// Skip current user if he has rated 0 common movies with $selectedUser
				if len(algorithms.Intersection[int](selectedUserMovies, userMovies)) == 0 {
					continue
				}
				// Finally, calculate the similarity using the requested similarity metric
				var similarity float64
				switch cfg.Similarity {
				case "jaccard":
					similarity = algorithms.JaccardSimilarity[int](selectedUserMovies, userMovies)
				case "dice":
					similarity = algorithms.DiceSimilarity[int](selectedUserMovies, userMovies)
				case "cosine":
					vectorA, vectorB := util.GetMovieRatingVectors(selectedUser, user, selectedUserMovies, userMovies)
					similarity = algorithms.CosineSimilarity[float32](vectorA, vectorB, algorithms.DotProductFloat32)
				case "pearson":
					vectorA, vectorB := util.GetMovieRatingVectors(selectedUser, user, selectedUserMovies, userMovies)
					similarity = algorithms.PearsonSimilarity[float32](vectorA, vectorB)
				}
				localSimilarUsers = append(localSimilarUsers, model.SimilarUser{
					UserID:     otherUserID,
					Similarity: similarity,
				})
			}
			// Keep the top-k most similar users this routine found
			updateMostSimilarUsers(&localSimilarUsers, cfg.K)
			// Append local results while protecting shared struct from concurrent writing
			mu.Lock()
			similarUsers = append(similarUsers, localSimilarUsers...)
			mu.Unlock()
		}(userChunk)
	}
	// Wait for all routines to finish
	wg.Wait()
	updateMostSimilarUsers(&similarUsers, cfg.K)
	return similarUsers
}

func updateMostSimilarUsers(users *[]model.SimilarUser, usersToKeep int) {
	// Sort similar movies by similarity in descending order and trim its size to usersToKeep
	sort.SliceStable((*users), func(i, j int) bool {
		return (*users)[i].Similarity > (*users)[j].Similarity
	})
	if len(*users) > usersToKeep {
		(*users) = (*users)[:usersToKeep]
	}
}
