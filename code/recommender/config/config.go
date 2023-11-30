package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

/*
Accepted values:
  - Algorithm: user, item, tag, hybrid
  - Similarity: jaccard, dice, cosine, pearson
  - Input: user_id, movie_id
*/
type Config struct {
	DataDir         string
	Recommendations int
	Similarity      string
	Algorithm       string
	Input           int
	// max{X} = -1 means read all
	MaxRecords int
	MaxUsers   int
	MaxTitles  int
	MaxMovies  int
	MaxTags    int
	WebServer  bool
}

func InitRecommender() (Config, error) {
	dataDir := flag.String("d", "", "Directory of data")
	numRecommendations := flag.Int("n", 0, "Number of recommendations")
	similarityMetric := flag.String("s", "", "Similarity metric")
	algorithm := flag.String("a", "", "Algorithm")
	input := flag.Int("i", 0, "Input")
	maxRecords := flag.Int("r", -1, "Max records to load")
	enableUI := flag.Bool("u", false, "Enable UI webserver")
	flag.Parse()

	var validationErrors []error
	usageMsg := fmt.Sprintln("\nUsage:\n" +
		"recommender -d directory_of_data -n number_of_recommendations -s similarity_metric -a algorithm -i input (-r maxRecordsToRead)\n" +
		"OR\n" +
		"recommender -d directory_of_data -u",
	)

	// Check if required flags are provided.
	if *dataDir == "" {
		return Config{}, errors.New(usageMsg)
	}
	// Check if the data directory exists.
	if _, err := os.Stat(*dataDir); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' directory does not exist", *dataDir)))
	}
	if len(*dataDir) == 0 || (*dataDir)[len(*dataDir)-1] != '/' {
		*dataDir += "/"
	}

	// Check if all necessary files exist
	ratingsFile := filepath.Join(*dataDir, "users.gob")
	if _, err := os.Stat(ratingsFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", ratingsFile)))
	}
	movieTitlesFile := filepath.Join(*dataDir, "movieTitles.gob")
	if _, err := os.Stat(movieTitlesFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", movieTitlesFile)))
	}
	moviesFile := filepath.Join(*dataDir, "movies.gob")
	if _, err := os.Stat(moviesFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", moviesFile)))
	}
	tagsFile := filepath.Join(*dataDir, "tags.gob")
	if _, err := os.Stat(tagsFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", tagsFile)))
	}

	if !*enableUI {
		// Check if required flags are provided.
		if *numRecommendations == 0 || *similarityMetric == "" || *algorithm == "" || *input == 0 {
			return Config{}, errors.New(usageMsg)
		}

		// Validate that provided similarity metric is accepted
		if *similarityMetric != "jaccard" && *similarityMetric != "dice" && *similarityMetric != "cosine" && *similarityMetric != "pearson" {
			validationErrors = append(validationErrors, errors.New("Allowed similarity metrics: 'jaccard', 'dice', 'cosine', 'pearson'"))
		}

		// Validate that provided algorithm is accepted
		if *algorithm != "user" && *algorithm != "item" && *algorithm != "tag" && *algorithm != "title" && *algorithm != "hybrid" {
			validationErrors = append(validationErrors, errors.New("Allowed similarity metrics: 'user', 'item', 'tag', 'title', 'hybrid'"))
		}

		// Validate that maxRecords is greater than 0 or -1 (default)
		if *maxRecords < 0 && *maxRecords != -1 {
			validationErrors = append(validationErrors, errors.New("Max records to load must be greater than 0 or -1"))
		}
	}

	// Check if any validation failed
	if len(validationErrors) > 0 {
		return Config{}, addToErrorList(validationErrors)
	}

	cfg := Config{
		DataDir:         *dataDir,
		Recommendations: *numRecommendations,
		Similarity:      *similarityMetric,
		Algorithm:       *algorithm,
		Input:           *input,
		MaxRecords:      *maxRecords,
		MaxUsers:        -1,
		MaxTitles:       -1,
		MaxMovies:       -1,
		MaxTags:         -1,
		WebServer:       *enableUI,
	}

	switch *algorithm {
	case "user":
		cfg.MaxUsers = *maxRecords
	case "item":
		cfg.MaxMovies = *maxRecords
	case "tag", "hybrid":
		cfg.MaxTags = *maxRecords
	case "title":
		cfg.MaxTitles = *maxRecords
	}

	return cfg, nil
}

func InitPreprocess() (Config, error) {
	dataDir := flag.String("d", "", "Directory of data")
	flag.Parse()

	var validationErrors []error

	// Check if required flags are provided.
	if *dataDir == "" {
		return Config{}, errors.New("Usage: preprocess -d directory_of_data")
	}
	if len(*dataDir) == 0 || (*dataDir)[len(*dataDir)-1] != '/' {
		*dataDir += "/"
	}

	// Check if the data directory exists.
	if _, err := os.Stat(*dataDir); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' directory does not exist", *dataDir)))
	}

	// Check if all necessary files exist
	ratingsFile := filepath.Join(*dataDir, "ratings.csv")
	if _, err := os.Stat(ratingsFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", ratingsFile)))
	}
	moviesFile := filepath.Join(*dataDir, "movies.csv")
	if _, err := os.Stat(moviesFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", moviesFile)))
	}
	tagsFile := filepath.Join(*dataDir, "tags.csv")
	if _, err := os.Stat(tagsFile); os.IsNotExist(err) {
		validationErrors = append(validationErrors, errors.New(fmt.Sprintf("'%s' was not found", tagsFile)))
	}

	// Check if any validation failed
	if len(validationErrors) > 0 {
		return Config{}, addToErrorList(validationErrors)
	}

	return Config{DataDir: *dataDir}, nil
}

func addToErrorList(errs []error) error {
	errStr := "Multiple validation errors occurred:\n"
	for _, err := range errs {
		errStr += "- " + err.Error() + "\n"
	}
	return errors.New(errStr)
}
