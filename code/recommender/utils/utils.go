package util

import (
	"fmt"
	"log"
	"os"
	"recommender/algorithms"
	model "recommender/models"
	"runtime/pprof"
	"strings"
)

/*
Split a given $set into chunks of $chunkSize Returns
a slice of chunks containing the initial $set data
*/
func GenerateChunkFromSet(set []int, chunkSize int) [][]int {
	if chunkSize <= 0 {
		chunkSize = 1
	}
	chunks := make([][]int, 0, (len(set)+chunkSize-1)/chunkSize)
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
*/
func GetMovieRatingVectors(user1 model.User, user2 model.User, movies1 []int, movies2 []int) ([]float32, []float32) {
	union := algorithms.Union[int](movies1, movies2)
	vectorA, vectorB := make([]float32, 0, len(union)), make([]float32, 0, len(union))
	for _, id := range union {
		if ratingA, exists := user1.MovieRatings[id]; exists {
			vectorA = append(vectorA, ratingA)
		} else {
			vectorA = append(vectorA, 0)
		}
		if ratingB, exists := user2.MovieRatings[id]; exists {
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
*/
func GetUserRatingVectors(movie1 model.Movie, movie2 model.Movie, users1 []int, users2 []int) ([]float32, []float32) {
	union := algorithms.Union[int](users1, users2)
	vectorA := make([]float32, 0, len(union))
	vectorB := make([]float32, 0, len(union))
	for _, id := range union {
		if ratingA, exists := movie1.UserRatings[id]; exists {
			vectorA = append(vectorA, ratingA)
		} else {
			vectorA = append(vectorA, 0)
		}
		if ratingB, exists := movie2.UserRatings[id]; exists {
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
*/
func GetTagOccurenceVectors(movie1TagOccurences map[string]int, movie2TagOccurences map[string]int) ([]int, []int) {
	movie1Tags, movie2Tags := []string{}, []string{}
	for tag := range movie1TagOccurences {
		movie1Tags = append(movie1Tags, tag)
	}
	for tag := range movie2TagOccurences {
		movie2Tags = append(movie2Tags, tag)
	}
	union := algorithms.Union[string](movie1Tags, movie2Tags)
	vectorA, vectorB := make([]int, 0, len(union)), make([]int, 0, len(union))
	for _, tag := range union {
		if occurrencesA, exists := movie1TagOccurences[tag]; exists {
			vectorA = append(vectorA, occurrencesA)
		} else {
			vectorA = append(vectorA, 0)
		}
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
	index := 0
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
			index++
		}
	}
	return vectorA, vectorB
}

/*
Create two boolean vectors based on two sets of elements, indicating the presence
or absence of each element in the respective sets. The final vectors have the same
size are aligned so that each vector[i] value refers to the same i (tag).
*/
func CreateBoolVectors[T comparable](set1 []T, set2 []T) (vectorA []bool, vectorB []bool) {
	union := algorithms.Union(set1, set2)
	// Create maps to check the presence of elements in set1 and set2.
	set1Map, set2Map := map[T]bool{}, map[T]bool{}
	for _, item := range set1 {
		set1Map[item] = true
	}
	for _, item := range set2 {
		set2Map[item] = true
	}
	// Populate vectors based on the presence of elements in union set.
	vectorA = make([]bool, 0, len(union))
	vectorB = make([]bool, 0, len(union))
	for _, item := range union {
		if set1Map[item] {
			vectorA = append(vectorA, true)
		} else {
			vectorA = append(vectorA, false)
		}
		if set2Map[item] {
			vectorB = append(vectorB, true)
		} else {
			vectorB = append(vectorB, false)
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
	// defer pprof.StopCPUProfile()
}

func StopProfiling() {
	pprof.StopCPUProfile()
}
