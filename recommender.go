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
	"recommender/recommenders"
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
	Message    string         `json:"message"`
	MetaInfo   string         `json:"metaInfo"`
}

type ResponseData struct {
	MovieID    int     `json:"movieID"`
	MovieTitle string  `json:"movieTitle"`
	Result     float64 `json:"result"`
}

var (
	// Number of routines to be spawned for parallelism
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
	// Indicator of whether the webserver is using a portion of the original dataset
	// This is useful to be able to reset the dataset to the original state after
	// a resuest has been made through the UI using the maxRecords functionality.
	limitedDataset = false
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
		// CLI mode
		var m runtime.MemStats
		startTime := time.Now()
		// Load only the files that are necessary for the selected algorithm to save time
		switch cfg.Algorithm {
		case "user":
			util.LoadData(&data.MovieTitles, cfg.DataDir+"movieTitles.gob", cfg.MaxTitles)
			util.LoadData(&data.Users, cfg.DataDir+"users.gob", cfg.MaxUsers)
		case "item":
			util.LoadData(&data.Movies, cfg.DataDir+"movies.gob", cfg.MaxMovies)
		case "tag":
			util.LoadData(&data.MovieTags, cfg.DataDir+"tags.gob", cfg.MaxTags)
		case "title":
			util.LoadData(&data.MovieTitles, cfg.DataDir+"movieTitles.gob", cfg.MaxTitles)
		case "hybrid":
			util.LoadData(&data.MovieTitles, cfg.DataDir+"movieTitles.gob", cfg.MaxTitles)
			util.LoadData(&data.Movies, cfg.DataDir+"movies.gob", cfg.MaxMovies)
			util.LoadData(&data.MovieTags, cfg.DataDir+"tags.gob", cfg.MaxTags)
		}
		err := checkRequestFeasibility(cfg.Algorithm, cfg.Input)
		if err != "" {
			fmt.Println(err)
			return
		}
		ratingForecasts, relevantMovies := performRecommendation(&cfg, &data)
		printRecommendations(&cfg, ratingForecasts, relevantMovies)
		fmt.Printf("Execution Time: %s\n", time.Since(startTime))
		runtime.ReadMemStats(&m)
		fmt.Printf("HeapAlloc: %d MiB\n", m.HeapAlloc/(1024*1024))
	}
}

func startWebServer(dataDir string) {
	fmt.Println("Starting Web-Server...")
	util.LoadData(&data.Users, dataDir+"users.gob")
	util.LoadData(&data.MovieTitles, dataDir+"movieTitles.gob")
	util.LoadData(&data.Movies, dataDir+"movies.gob")
	util.LoadData(&data.MovieTags, dataDir+"tags.gob")
	// Register API endpoint handlers
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(os.Getenv("PWD")+"/ui"))))
	http.HandleFunc("/recommend", func(w http.ResponseWriter, r *http.Request) {
		handleRecommendationRequest(w, r, dataDir)
	})
	// Start the Web-Server
	fmt.Println("Web-Server UI is available on http://localhost:8080/ui/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Main handler of the Web-Server to
func handleRecommendationRequest(w http.ResponseWriter, r *http.Request, dataDir string) {
	startTime := time.Now()
	// Prepare default response
	response := ResponseTemplate{
		Status:     "success",
		StatusCode: 200,
		Data:       []ResponseData{},
		Message:    "",
	}
	// Retrieve & parse query parameters. Also update the dataset if necessary
	queryParams := r.URL.Query()
	recommendations, _ := strconv.Atoi(queryParams["recommendations"][0])
	similarity := queryParams["similarity"][0]
	algorithm := queryParams["algorithm"][0]
	input, _ := strconv.Atoi(queryParams["input"][0])
	maxRecords := -1
	if _, exists := queryParams["maxRecords"]; exists {
		maxRecords, _ = strconv.Atoi(queryParams["maxRecords"][0])
		reloadData(algorithm, maxRecords, dataDir)
		limitedDataset = true
	}
	// Create a custom configuration object based on query params to perform recommendation
	cfg := config.Config{
		Recommendations: recommendations,
		Similarity:      similarity,
		Algorithm:       algorithm,
		Input:           input,
		MaxRecords:      maxRecords,
	}
	fmt.Printf("Received request with parameters: -n=%d -s=%s -a=%s -i=%d -r=%d\n",
		recommendations, similarity, algorithm, input, maxRecords)
	err := checkRequestFeasibility(cfg.Algorithm, cfg.Input)
	if err == "" {
		// Request is feasible, proceed to recommendation
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
		} else if len(relevantMovies) != 0 {
			for _, relevantMovie := range relevantMovies {
				response.Data = append(response.Data, ResponseData{
					MovieID:    relevantMovie.MovieID,
					MovieTitle: data.MovieTitles[relevantMovie.MovieID].Title,
					Result:     math.Trunc((relevantMovie.Similarity * 100000)) / 100000,
				})
			}
			// Additional info for the requested movie
			response.MetaInfo = data.MovieTitles[cfg.Input].Title
		} else {
			response.Message = fmt.Sprintf("No relevant movies found for user %d. Try using another algorithm.", cfg.Input)
		}
	} else {
		response.Message = err
		fmt.Println("Request is not feasible:", err)
	}
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	fmt.Printf("Reponse sent in: %s\n", time.Since(startTime))
	// Reset dataset to the original state
	if limitedDataset {
		reloadData(algorithm, -1, dataDir)
		limitedDataset = false
	}
}

// Reload data mechanism in case the user requests limited dataset through the UI
func reloadData(algorithm string, maxRecords int, dataDir string) {
	switch algorithm {
	case "user":
		util.LoadData(&data.Users, dataDir+"users.gob", maxRecords)
	case "item", "hybrid":
		util.LoadData(&data.Movies, dataDir+"movies.gob", maxRecords)
	case "tag":
		util.LoadData(&data.MovieTags, dataDir+"tags.gob", maxRecords)
	case "title":
		util.LoadData(&data.MovieTitles, dataDir+"movieTitles.gob", maxRecords)
	}
}

// The core function of the recommender both when using the CLI or the UI interface
func performRecommendation(cfg *config.Config, data *Data) ([]model.Rating, []model.SimilarMovie) {
	ratingForecasts := make([]model.Rating, 0, cfg.Recommendations)
	relevantMovies := make([]model.SimilarMovie, 0, cfg.Recommendations)
	switch cfg.Algorithm {
	case "user":
		ratingForecasts = recommenders.RecommendBasedOnUser(cfg, &data.Users, &data.MovieTitles)
	case "item":
		ratingForecasts = recommenders.RecommendBasedOnItem(cfg, &data.Movies)
	case "tag":
		relevantMovies = recommenders.RecommendBasedOnTag(cfg, &data.MovieTags)
	case "title":
		relevantMovies = recommenders.RecommendBasedOnTitle(cfg, &data.MovieTitles)
	case "hybrid":
		relevantMovies = recommenders.RecommendHybrid(cfg, &data.MovieTitles, &data.Movies, &data.MovieTags)
	}
	return ratingForecasts, relevantMovies
}

// Prints results if running in CLI mode
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
	case "tag", "title", "hybrid":
		if len(relevantMovies) == 0 {
			fmt.Printf("No relevant movies found for movie %d. Try using another algorithm.\n", cfg.Input)
			break
		}
		fmt.Printf("Top movie recommendations for movie %d '%s' are:\n", cfg.Input, movieTitles[cfg.Input].Title)
		for i, recommendation := range relevantMovies {
			fmt.Printf("%d: ID: %d, Title: %s => %.5f\n", i+1, recommendation.MovieID, movieTitles[recommendation.MovieID].Title, recommendation.Similarity)
		}
	}
}

// Checks if the request can be satisfied for the given input.
// Returns empty string if request is feasible or an error message if not.
func checkRequestFeasibility(algorithm string, input int) string {
	switch algorithm {
	case "user":
		if _, exists := data.Users[input]; !exists {
			return "User ID not found in current dataset. Please try with another ID."
		}
	case "item":
		if _, exists := data.Movies[input]; !exists {
			return "Movie ID not found in current dataset. Please try with another ID."
		}
	case "tag":
		if _, exists := data.MovieTags[input]; !exists {
			return "Movie ID not found in current dataset. Please try with another ID."
		}
	case "title":
		if _, exists := data.MovieTitles[input]; !exists {
			return "Movie ID not found in current dataset. Please try with another ID."
		}
	case "hybrid":
		if _, exists := data.Movies[input]; !exists {
			return "Movie ID not found in current dataset. Please try with another ID."
		}
	}
	return ""
}
