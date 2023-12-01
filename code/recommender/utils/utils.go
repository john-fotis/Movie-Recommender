package util

import (
	"fmt"
	"log"
	"os"
	model "recommender/models"
	"runtime/pprof"
	"strings"
)

// Returns a slice of $numThread chunks (slices) containing the initial $set data
func GenerateChunkFromSet(set []int, numThreads int) [][]int {
	if numThreads <= 0 {
		numThreads = 1
	}
	chunkSize := (len(set) + numThreads - 1) / numThreads
	chunks := make([][]int, 0, numThreads)
	for i := 0; i < len(set); i += chunkSize {
		end := i + chunkSize
		if end > len(set) {
			end = len(set)
		}
		chunks = append(chunks, set[i:end])
	}
	return chunks
}

/*
Generate 2 vectors which contain movie ratings based on a slice of
movie ratings for each user. The final vectors have the same size
and are aligned so that each vector[i] value refers to the same i (user).
Notice: Priority is given to movies1, which means after all IDs in movies1 are processed, the
rest IDs in movies2 (if any) will be ignored. Final vectors length is equal to len(movies1).
*/
func GetMovieRatingVectors(user1 model.User, user2 model.User, movies1 []int, movies2 []int) ([]float32, []float32) {
	vectorA, vectorB := make([]float32, 0, len(movies1)), make([]float32, 0, len(movies1))
	for _, movieID := range movies1 {
		vectorA = append(vectorA, user1.MovieRatings[movieID])
		if ratingB, exists := user2.MovieRatings[movieID]; exists {
			vectorB = append(vectorB, ratingB)
		} else {
			vectorB = append(vectorB, 0)
		}
	}
	return vectorA, vectorB
}

/*
Generate 2 vectors which contain user ratings based on a slice of user
ratings for each movie. The final vectors have the same size and are
aligned so that each vector[i] value refers to the same i (movie).
Notice: Priority is given to users1, which means after all IDs in users1 are processed, the
rest IDs in users2 (if any) will be ignored. Final vectors length is equal to len(users1).
*/
func GetUserRatingVectors(movie1 model.Movie, movie2 model.Movie, users1 []int, users2 []int) ([]float32, []float32) {
	vectorA, vectorB := make([]float32, 0, len(users1)), make([]float32, 0, len(users1))
	for _, userID := range users1 {
		vectorA = append(vectorA, movie1.UserRatings[userID])
		if ratingB, exists := movie2.UserRatings[userID]; exists {
			vectorB = append(vectorB, ratingB)
		} else {
			vectorB = append(vectorB, 0)
		}
	}
	return vectorA, vectorB
}

/*
Generate 2 vectors which contain tag occurrences based on a map of
{tag:occurrences} pairs for each vector. The final vectors have the same
size and are aligned so that each vector[i] value refers to the same i (tag).
Notice: Priority is given to movie1, which means after all tags of movie1 are processed, the rest
tags of movie2 (if any) will be ignored. Final vectors length is equal to len(movie1TagOccurences).
*/
func GetTagOccurenceVectors(movie1TagOccurences map[string]int, movie2TagOccurences map[string]int) ([]int, []int) {
	movie1Tags, movie2Tags := []string{}, []string{}
	for tag := range movie1TagOccurences {
		movie1Tags = append(movie1Tags, tag)
	}
	for tag := range movie2TagOccurences {
		movie2Tags = append(movie2Tags, tag)
	}
	vectorA, vectorB := make([]int, 0, len(movie1Tags)), make([]int, 0, len(movie1Tags))
	for _, tag := range movie1Tags {
		vectorA = append(vectorA, movie1TagOccurences[tag])
		if occurrencesB, exists := movie2TagOccurences[tag]; exists {
			vectorB = append(vectorB, occurrencesB)
		} else {
			vectorB = append(vectorB, 0)
		}
	}
	return vectorA, vectorB
}

/*
Generate 2 vectors which contain the results of TF.IDF multiplication based on an IDF
map consisted of {token:IDFscore} pairs and two TF maps consisted of {token:TFscore}.
*/
func GetTfIdfVectors(idfMap map[string]float64, tfMap1, tfMap2 map[string]float64) ([]float64, []float64) {
	vectorA, vectorB := make([]float64, 0), make([]float64, 0)
	for token, idf := range idfMap {
		_, exists1 := tfMap1[token]
		_, exists2 := tfMap2[token]
		// If token doesn't exist in any of the TF maps skip it, no need to keep pairs of zeros
		if exists1 || exists2 {
			if exists1 {
				vectorA = append(vectorA, tfMap1[token]*idf)
			} else {
				vectorA = append(vectorA, 0)
			}
			if exists2 {
				vectorB = append(vectorB, tfMap2[token]*idf)
			} else {
				vectorB = append(vectorB, 0)
			}
		}
	}
	return vectorA, vectorB
}

/*
Create two boolean vectors based on two sets of elements, indicating the presence
or absence of each element in the respective sets. The final vectors have the same
size are aligned so that each vector[i] value refers to the same i (tag).
Notice: Priority is given to set1, which means after all items of set1 are processed, the rest
items of set2 (if any) will be ignored. Final vectors length is equal to len(set1).
*/
func CreateBoolVectors[T comparable](set1 []T, set2 []T) (vectorA []bool, vectorB []bool) {
	// Populate vectors based on the presence of elements in union set.
	vectorA, vectorB = make([]bool, 0, len(set1)), make([]bool, 0, len(set1))
	for _, item1 := range set1 {
		vectorA = append(vectorA, true)
		vectorB = append(vectorB, false)
		for _, item2 := range set2 {
			if item1 == item2 {
				vectorB[len(vectorB)-1] = true
				break
			}
		}
	}
	return vectorA, vectorB
}

func StartProfiling(fileName string) {
	profilingDir := os.Getenv("PWD") + "/profiling/"
	err := os.MkdirAll(profilingDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
		return
	}
	if !strings.HasSuffix(fileName, ".prof") {
		fileName += ".prof"
	}
	file, err := os.Create(profilingDir + fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	pprof.StartCPUProfile(file)
}

func StopProfiling() {
	pprof.StopCPUProfile()
}
