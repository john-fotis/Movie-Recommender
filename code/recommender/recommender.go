package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"runtime"
	"strconv"
	"time"
)

type Data struct {
	Users       map[int]model.User
	MovieTitles map[int]model.MovieTitle
	Movies      map[int]model.Movie
	MovieTags   map[int]model.MovieTags
}

type ResponseTemplate struct {
	Status     string         `json:"status"`
	StatusCode int            `json:"statusCode"`
	Data       []ResponseData `json:"data"`
	DataType   string         `json:"dataType"`
	Message    string         `json:"message"`
}

type ResponseData struct {
	MovieID    int     `json:"movieID"`
	MovieTitle string  `json:"movieTitle"`
	Result     float64 `json:"result"`
}

var (
	// Number of threads to implement parallelism
	numThreads = 8
	// Set size to be used
	k = 128
	// Main struct to store data
	data = Data{
		Users:       make(map[int]model.User, 0),
		MovieTitles: make(map[int]model.MovieTitle, 0),
		Movies:      make(map[int]model.Movie, 0),
		MovieTags:   make(map[int]model.MovieTags, 0),
	}
)

func main() {
	cfg, err := config.InitRecommender()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
		return
	}
	if cfg.WebServer {
		startWebServer(cfg.DataDir)
	} else {
		var m runtime.MemStats
		startTime := time.Now()
		util.LoadData(&data.MovieTitles, cfg.DataDir+"movieTitles.gob", cfg.MaxTitles)
		switch cfg.Algorithm {
		case "user":
			util.LoadData(&data.Users, cfg.DataDir+"users.gob", cfg.MaxUsers)
		case "item":
			util.LoadData(&data.Movies, cfg.DataDir+"movies.gob", cfg.MaxMovies)
		case "tag":
			util.LoadData(&data.MovieTags, cfg.DataDir+"tags.gob", cfg.MaxTags)
		case "hybrid":
			util.LoadData(&data.Users, cfg.DataDir+"users.gob")
			util.LoadData(&data.MovieTags, cfg.DataDir+"tags.gob", cfg.MaxTags)
		}
		ratingForecasts, relevantMovies := performRecommendation(&cfg, &data)
		printRecommendations(&cfg, ratingForecasts, relevantMovies)
		fmt.Printf("Execution Time: %s\n", time.Since(startTime))
		runtime.ReadMemStats(&m)
		fmt.Printf("HeapAlloc: %d MiB\n", m.HeapAlloc/(1024*1024))
	}
}

func startWebServer(dataDir string) {
	util.LoadData(&data.Users, dataDir+"users.gob")
	util.LoadData(&data.MovieTitles, dataDir+"movieTitles.gob")
	util.LoadData(&data.Movies, dataDir+"movies.gob")
	util.LoadData(&data.MovieTags, dataDir+"tags.gob")

	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(os.Getenv("PWD")+"/ui"))))
	http.HandleFunc("/recommend", handleWrapper(dataDir))

	fmt.Printf("Starting Web-Server on port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleWrapper(dataDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleRecommendationRequest(w, r, dataDir)
	}
}

func handleRecommendationRequest(w http.ResponseWriter, r *http.Request, dataDir string) {
	// Prepare default response
	response := ResponseTemplate{
		Status:     "success",
		StatusCode: 200,
		Data:       []ResponseData{},
		DataType:   "",
		Message:    "",
	}
	// Retrieve & parse query parameters, and reaload Data if necessary
	queryParams := r.URL.Query()
	recommendations, _ := strconv.Atoi(queryParams["recommendations"][0])
	similarity := queryParams["similarity"][0]
	algorithm := queryParams["algorithm"][0]
	input, _ := strconv.Atoi(queryParams["input"][0])
	maxRecords := -1
	if _, exists := queryParams["maxRecords"]; exists {
		maxRecords, _ = strconv.Atoi(queryParams["maxRecords"][0])
		reloadData(algorithm, maxRecords, dataDir)
	}
	// Create a custom configurationobject based on query params to perform recommendation
	cfg := config.Config{
		Recommendations: recommendations,
		Similarity:      similarity,
		Algorithm:       algorithm,
		Input:           input,
		MaxRecords:      maxRecords,
	}
	// Proceed to recommendation
	ratingForecasts, relevantMovies := performRecommendation(&cfg, &data)
	// Fill the response content based on the type of the recommendation results
	if len(ratingForecasts) != 0 {
		for _, movieRating := range ratingForecasts {
			response.Data = append(response.Data, ResponseData{
				MovieID:    movieRating.MovieID,
				MovieTitle: data.MovieTitles[movieRating.MovieID].Title,
				Result:     math.Trunc((float64(movieRating.Rating) * 100)) / 100,
			})
		}
		response.DataType = "ratings"
	} else if len(relevantMovies) != 0 {
		for _, relevantMovie := range relevantMovies {
			response.Data = append(response.Data, ResponseData{
				MovieID:    relevantMovie.MovieID,
				MovieTitle: data.MovieTitles[relevantMovie.MovieID].Title,
				Result:     math.Trunc((relevantMovie.Similarity * 100000)) / 100000,
			})
		}
		response.DataType = "similarities"
	} else {
		response.Message = fmt.Sprintf("No relevant movies found for user %d. Try using another algorithm.", cfg.Input)
	}
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func reloadData(algorithm string, maxRecords int, dataDir string) {
	switch algorithm {
	case "user":
		util.LoadData(&data.Users, dataDir+"users.gob", maxRecords)
	case "item":
		util.LoadData(&data.Movies, dataDir+"movies.gob", maxRecords)
	case "tag":
		util.LoadData(&data.MovieTags, dataDir+"tags.gob", maxRecords)
	case "hybrid":
		util.LoadData(&data.MovieTags, dataDir+"tags.gob", maxRecords)
	}
}

func performRecommendation(cfg *config.Config, data *Data) ([]model.Rating, []model.SimilarMovie) {
	ratingForecasts := make([]model.Rating, 0, cfg.Recommendations)
	relevantMovies := make([]model.SimilarMovie, 0, len(data.MovieTitles))
	switch cfg.Algorithm {
	case "user":
		if _, exists := data.Users[cfg.Input]; !exists {
			fmt.Println("User ID not found in current dataset. Please try with another ID.")
			return nil, nil
		}
		ratingForecasts = RecommendBasedOnUser(cfg, &data.Users, &data.MovieTitles)
	case "item":
		if _, exists := data.Movies[cfg.Input]; !exists {
			fmt.Println("Movie ID not found in current dataset. Please try with another ID.")
			return nil, nil
		}
		ratingForecasts = RecommendBasedOnItem(cfg, &data.Movies, cfg.Input)
	case "tag":
		if _, exists := data.MovieTags[cfg.Input]; !exists {
			fmt.Println("Movie ID not found in current dataset. Please try with another ID.")
			return nil, nil
		}
		relevantMovies = RecommendBasedOnTag(cfg, &data.MovieTags)
	case "title":
		if _, exists := data.MovieTitles[cfg.Input]; !exists {
			fmt.Println("Movie ID not found in current dataset. Please try with another ID.")
			return nil, nil
		}
		relevantMovies = RecommendBasedOnTitle(cfg, &data.MovieTitles)
	case "hybrid":
		if _, exists := data.Users[cfg.Input]; !exists {
			fmt.Println("User ID not found in current dataset. Please try with another ID.")
			return nil, nil
		}
		selectedUserRatings := data.Users[cfg.Input].MovieRatings
		if len(selectedUserRatings) < medianRatedMoviesPerUser(&data.Users) {
			// If the user has rated more movies than the median of all users we consider that his
			// taste in movies can be estimated accuratelly and thus we apply user-user similarity
			relevantMovies = RecommendHybrid(cfg, selectedUserRatings, &data.MovieTitles, &data.Movies, &data.MovieTags)
		} else {
			// If the user has rated relatively little movies (eg. new user) title and tag based
			// recommendation on the movies he has rated will give more accurate results
			ratingForecasts = RecommendBasedOnUser(cfg, &data.Users, &data.MovieTitles)
		}
	}
	return ratingForecasts, relevantMovies
}

func printRecommendations(cfg *config.Config, ratingForecasts []model.Rating, relevantMovies []model.SimilarMovie) {
	movieTitles := make(map[int]model.MovieTitle)
	util.LoadData(&movieTitles, cfg.DataDir+"movieTitles.gob")
	switch cfg.Algorithm {
	case "user", "item":
		if len(ratingForecasts) == 0 {
			fmt.Printf("No movie recommendations for user %d. Try using another algorithm.\n", cfg.Input)
			break
		}
		fmt.Printf("Top movie recommendations for user %d are:\n", cfg.Input)
		for i, recommendation := range ratingForecasts {
			fmt.Printf("%d: ID: %d, Title: %s => %.2f\n",
				i+1, recommendation.MovieID, movieTitles[recommendation.MovieID].Title, recommendation.Rating,
			)
		}
	case "tag", "title":
		if len(relevantMovies) == 0 {
			fmt.Printf("No relevant movies found for movie %d. Try using another algorithm.\n", cfg.Input)
			break
		}
		fmt.Printf("Top movie recommendations for movie %d are:\n", cfg.Input)
		for i, recommendation := range relevantMovies {
			fmt.Printf("%d: ID: %d, Title: %s => %.2f\n", i+1, recommendation.MovieID, movieTitles[recommendation.MovieID].Title, recommendation.Similarity)
		}
	case "hybrid":
		if len(ratingForecasts) != 0 {
			for i, recommendation := range ratingForecasts {
				fmt.Printf("%d: ID: %d, Title: %s => %.2f\n", i+1, recommendation.MovieID, movieTitles[recommendation.MovieID].Title, recommendation.Rating)
			}
		} else if len(relevantMovies) != 0 {
			for i, recommendation := range relevantMovies {
				fmt.Printf("%d: ID: %d, Title: %s => %.5f\n", i+1, recommendation.MovieID, movieTitles[recommendation.MovieID].Title, recommendation.Similarity)
			}
		} else {
			fmt.Printf("No relevant movies found for user %d. Try using another algorithm.\n", cfg.Input)
		}
	}
}
