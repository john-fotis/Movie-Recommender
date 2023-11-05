package config

import (
	"errors"
	"flag"
)

type Config struct {
	DataDir         string
	Recommendations int
	Similarity      string
	Algorithm       string
	Input           int
}

func InitServer() (Config, error) {
	dataDir := flag.String("d", "", "Directory of data")
	numRecommendations := flag.Int("n", 0, "Number of recommendations")
	similarityMetric := flag.String("s", "", "Similarity metric")
	algorithm := flag.String("a", "", "Algorithm")
	input := flag.Int("i", 0, "Input")
	flag.Parse()

	// Check if required flags are provided.
	if *dataDir == "" || *numRecommendations == 0 || *similarityMetric == "" || *algorithm == "" || *input == 0 {
		return Config{}, errors.New("Usage: recommender -d directory_of_data -n number_of_recommendations -s similarity_metric -a algorithm -i input")
	}

	cfg := Config{
		DataDir:         *dataDir,
		Recommendations: *numRecommendations,
		Similarity:      *similarityMetric,
		Algorithm:       *algorithm,
		Input:           *input,
	}

	return cfg, nil
}
