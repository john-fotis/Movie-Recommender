package main

import (
	"fmt"
	"log"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"runtime"
	"time"
)

// Data represents the data structures for different CSV files.
type Data struct {
	Users       map[int]model.User
	MovieTitles map[int]model.MovieTitle
	Movies      map[int]model.Movie
	MovieTags   map[int]model.MovieTags
}

var (
	// Number of threads to implement parallelism
	numThreads = 8
	// Set size to be used
	k = 128
)

func main() {
	var m runtime.MemStats
	startTime := time.Now()

	cfg, err := config.InitServer()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
		return
	}
	data := Data{
		Users:       make(map[int]model.User, 0),
		MovieTitles: make(map[int]model.MovieTitle, 0),
		Movies:      make(map[int]model.Movie, 0),
		MovieTags:   make(map[int]model.MovieTags, 0),
	}

	util.LoadData(&data.MovieTitles, cfg.DataDir+"movieTitles.gob")

	ratingForecasts := make([]model.Rating, 0, cfg.Recommendations)
	relevantMovies := make([]model.SimilarMovie, 0, len(data.MovieTitles))
	switch cfg.Algorithm {
	case "user":
		util.LoadData(&data.Users, cfg.DataDir+"users.gob", cfg.MaxUsers)
		if _, exists := data.Users[cfg.Input]; !exists {
			fmt.Println("User ID not found in current dataset. Please try with another ID.")
			return
		}
		ratingForecasts = RecommendBasedOnUser(&cfg, &data.Users, &data.MovieTitles)
	case "item":
		util.LoadData(&data.Movies, cfg.DataDir+"movies.gob", cfg.MaxMovies)
		if _, exists := data.Movies[cfg.Input]; !exists {
			fmt.Println("Movie ID not found in current dataset. Please try with another ID.")
			return
		}
		ratingForecasts = RecommendBasedOnItem(&cfg, &data.Movies, cfg.Input)
	case "tag":
		util.LoadData(&data.MovieTags, cfg.DataDir+"tags.gob", cfg.MaxTags)
		if _, exists := data.MovieTags[cfg.Input]; !exists {
			fmt.Println("Movie ID not found in current dataset. Please try with another ID.")
			return
		}
		relevantMovies = RecommendBasedOnTag(&cfg, &data.MovieTags)
	case "title":
		if _, exists := data.MovieTitles[cfg.Input]; !exists {
			fmt.Println("Movie ID not found in current dataset. Please try with another ID.")
			return
		}
		relevantMovies = RecommendBasedOnTitle(&cfg, &data.MovieTitles)
	case "hybrid":
		util.LoadData(&data.Users, cfg.DataDir+"users.gob", cfg.MaxMovies)
		if _, exists := data.Users[cfg.Input]; !exists {
			fmt.Println("User ID not found in current dataset. Please try with another ID.")
			return
		}
		selectedUserRatings := data.Users[cfg.Input].MovieRatings
		if len(selectedUserRatings) < medianRatedMoviesPerUser(&data.Users) {
			// If the user has rated more movies than the median of all users we consider that his
			// taste in movies can be estimated accuratelly and thus we apply user-user similarity
			util.LoadData(&data.MovieTags, cfg.DataDir+"tags.gob", cfg.MaxTags)
			// Free up memory used for users since we wont need it in this case
			data.Users = nil
			relevantMovies = RecommendHybrid(&cfg, selectedUserRatings, &data.MovieTitles, &data.Movies, &data.MovieTags)
		} else {
			// If the user has rated relatively little movies (eg. new user) title and tag based
			// recommendation on the movies he has rated will give more accurate results
			ratingForecasts = RecommendBasedOnUser(&cfg, &data.Users, &data.MovieTitles)
		}
	}
	// Print recommendations
	switch cfg.Algorithm {
	case "user", "item":
		if len(ratingForecasts) == 0 {
			fmt.Printf("No movie recommendations for user %d. Try using another algorithm.\n", cfg.Input)
			break
		}
		fmt.Printf("Top movie recommendations for user %d are:\n", cfg.Input)
		for i, recommendation := range ratingForecasts {
			fmt.Printf("%d: ID: %d, Title: %s => %.2f\n",
				i+1, recommendation.MovieID, data.MovieTitles[recommendation.MovieID].Title, recommendation.Rating,
			)
		}
	case "tag", "title":
		if len(relevantMovies) == 0 {
			fmt.Printf("No relevant movies found for movie %d. Try using another algorithm.\n", cfg.Input)
			break
		}
		fmt.Printf("Top movie recommendations for movie %d are:\n", cfg.Input)
		for i, recommendation := range relevantMovies {
			fmt.Printf("%d: ID: %d, Title: %s => %.2f\n", i+1, recommendation.MovieID, data.MovieTitles[recommendation.MovieID].Title, recommendation.Similarity)
		}
	case "hybrid":
		if len(ratingForecasts) != 0 {
			for i, recommendation := range ratingForecasts {
				fmt.Printf("%d: ID: %d, Title: %s => %.2f\n", i+1, recommendation.MovieID, data.MovieTitles[recommendation.MovieID].Title, recommendation.Rating)
			}
		} else if len(relevantMovies) != 0 {
			for i, recommendation := range relevantMovies {
				fmt.Printf("%d: ID: %d, Title: %s => %.2f\n", i+1, recommendation.MovieID, data.MovieTitles[recommendation.MovieID].Title, recommendation.Similarity)
			}
		} else {
			fmt.Printf("No relevant movies found for user %d. Try using another algorithm.\n", cfg.Input)
		}
	}
	fmt.Printf("Execution Time: %s\n", time.Since(startTime))
	runtime.ReadMemStats(&m)
	fmt.Printf("HeapAlloc: %d MiB\n", m.HeapAlloc/(1024*1024))
}
