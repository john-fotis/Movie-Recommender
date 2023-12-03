package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"strings"
)

func main() {
	cfg, err := config.InitPreprocess()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
		return
	}

	preprocessedFilesDir := os.Getenv("PWD") + "/preprocess/data/"
	err = os.MkdirAll(preprocessedFilesDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
		return
	}

	movieTitles := make(map[int]model.MovieTitle)
	util.LoadCSVData(&movieTitles, cfg.DataDir+"movies.csv")
	writeGOBToFile(movieTitles, preprocessedFilesDir+"movieTitles.gob")

	users := make(map[int]model.User)
	util.LoadCSVData(&users, cfg.DataDir+"ratings.csv")
	writeGOBToFile(users, preprocessedFilesDir+"users.gob")

	movies := make(map[int]model.Movie)
	util.LoadCSVData(&movies, cfg.DataDir+"ratings.csv")
	writeGOBToFile(movies, preprocessedFilesDir+"movies.gob")

	tags := make(map[int]model.MovieTags)
	util.LoadCSVData(&tags, cfg.DataDir+"tags.csv")
	writeGOBToFile(tags, preprocessedFilesDir+"tags.gob")
}

// Stores a data interface into a file using Go Binary format
func writeGOBToFile(data interface{}, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Failed to create file: %s\n", err)
		return
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		fmt.Printf("Failed to encode GOB: %s\n", err)
		return
	}
	pathTokens := strings.Split(filePath, "/")
	fileName := strings.Trim(pathTokens[len(pathTokens)-1], ".gob")
	fileName = strings.ToUpper(fileName[:1]) + fileName[1:]
	fmt.Printf("%s data encoded and written to file: %s\n", fileName, filePath)
}
