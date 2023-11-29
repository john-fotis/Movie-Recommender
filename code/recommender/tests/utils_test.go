package tests

import (
	model "recommender/models"
	util "recommender/utils"
	"reflect"
	"testing"
)

func TestGenerateChunkFromSet(t *testing.T) {
	inputSet := []int{1, 2, 3, 4, 5, 6, 7, 8}
	chunkSize := 3
	expectedChunks := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8},
	}
	chunks := util.GenerateChunkFromSet(inputSet, chunkSize)
	if !reflect.DeepEqual(chunks, expectedChunks) {
		t.Errorf("Generated chunks do not match the expected result. Got: %v, Expected: %v", chunks, expectedChunks)
	}
}

func TestGetMovieRatingVectors(t *testing.T) {
	user1 := model.User{
		MovieRatings: map[int]float32{
			1: 4.0,
			2: 3.5,
			3: 5.0,
		},
	}
	user2 := model.User{
		MovieRatings: map[int]float32{
			2: 4.5,
			3: 3.0,
			4: 2.5,
		},
	}
	movies1 := []int{1, 2, 3}
	movies2 := []int{2, 3, 4}
	expectedVectorA := []float32{4.0, 3.5, 5.0, 0.0}
	expectedVectorB := []float32{0.0, 4.5, 3.0, 2.5}
	vectorA, vectorB := util.GetMovieRatingVectors(user1, user2, movies1, movies2)
	if len(vectorA) != len(vectorB) || len(vectorA) != len(expectedVectorA) {
		t.Errorf("Vector A or Vector B not of the expected size.")
		return
	}
	if !reflect.DeepEqual(vectorA, expectedVectorA) {
		t.Errorf("Vector A does not match the expected result. Got: %v, Expected: %v", vectorA, expectedVectorA)
	}
	if !reflect.DeepEqual(vectorB, expectedVectorB) {
		t.Errorf("Vector B does not match the expected result. Got: %v, Expected: %v", vectorB, expectedVectorB)
	}
}

func TestGetUserRatingVectors(t *testing.T) {
	movie1 := model.Movie{
		UserRatings: map[int]float32{
			1: 4.0,
			2: 3.5,
			3: 5.0,
		},
	}
	movie2 := model.Movie{
		UserRatings: map[int]float32{
			2: 4.5,
			3: 3.0,
			4: 2.5,
		},
	}
	users1 := []int{1, 2, 3}
	users2 := []int{2, 3, 4}
	expectedVectorA := []float32{4.0, 3.5, 5.0, 0.0}
	expectedVectorB := []float32{0.0, 4.5, 3.0, 2.5}
	vectorA, vectorB := util.GetUserRatingVectors(movie1, movie2, users1, users2)
	if len(vectorA) != len(vectorB) || len(vectorA) != len(expectedVectorA) {
		t.Errorf("Vector A or Vector B not of the expected size.")
		return
	}
	if !reflect.DeepEqual(vectorA, expectedVectorA) {
		t.Errorf("Vector A does not match the expected result. Got: %v, Expected: %v", vectorA, expectedVectorA)
	}
	if !reflect.DeepEqual(vectorB, expectedVectorB) {
		t.Errorf("Vector B does not match the expected result. Got: %v, Expected: %v", vectorB, expectedVectorB)
	}
}

func TestGetTagOccurenceVectors(t *testing.T) {
	movie1TagOccurrences := map[string]int{
		"tag1": 3,
		"tag2": 2,
		"tag3": 4,
	}
	movie2TagOccurrences := map[string]int{
		"tag2": 1,
		"tag3": 5,
		"tag4": 2,
	}
	expectedVectorA := []int{3, 2, 4, 0}
	expectedVectorB := []int{0, 1, 5, 2}
	vectorA, vectorB := util.GetTagOccurenceVectors(movie1TagOccurrences, movie2TagOccurrences)
	if len(vectorA) != len(vectorB) || len(vectorA) != len(expectedVectorA) {
		t.Errorf("Vector A or Vector B not of the expected size.")
		return
	}
	vectorsDontMatch := false
	for i := range vectorA {
		foundExpectedMatch := false
		for j := range expectedVectorA {
			if vectorA[i] == expectedVectorA[j] && vectorB[i] == expectedVectorB[j] {
				foundExpectedMatch = true
				break
			}
		}
		// If at least one pair of vectorA[i], vectorB[i] doesn't match any pair of
		// expectedVectorA[j], expectedVectorB[j] then this pair is wrong.
		if !foundExpectedMatch {
			vectorsDontMatch = true
			break
		}
	}
	if vectorsDontMatch {
		t.Error("Vectors are not as expected.")
		t.Errorf("Got: VectorA=%v, VectorB=%v", vectorA, vectorB)
		t.Errorf("Expected: VectorA=%v, VectorB=%v", expectedVectorA, expectedVectorB)
	}
}

func TestCreateBoolVectors(t *testing.T) {
	set1 := []string{"a", "b", "c"}
	set2 := []string{"b", "c", "d"}
	expectedVectorA := []bool{true, true, true, false}
	expectedVectorB := []bool{false, true, true, true}
	vectorA, vectorB := util.CreateBoolVectors(set1, set2)
	if len(vectorA) != len(vectorB) || len(vectorA) != len(expectedVectorA) {
		t.Errorf("Vector A or Vector B not of the expected size.")
		return
	}
	if !reflect.DeepEqual(vectorA, expectedVectorA) {
		t.Errorf("Vector A does not match the expected result. Got: %v, Expected: %v", vectorA, expectedVectorA)
	}
	if !reflect.DeepEqual(vectorB, expectedVectorB) {
		t.Errorf("Vector B does not match the expected result. Got: %v, Expected: %v", vectorB, expectedVectorB)
	}
}

func TestGetTfIdfVectors(t *testing.T) {
	idfMap := map[string]float64{
		"word1": 0.2,
		"word2": 0.04,
		"word3": 0.1,
		"word4": 0.1,
		"word5": 0.3,
	}
	tfMap1 := map[string]float64{
		"word2": 0.25,
		"word1": 0.25,
		"word4": 0.5,
	}
	tfMap2 := map[string]float64{
		"word4": 0.5,
		"word5": 0.5,
	}
	expectedVectorA := []float64{0.05, 0.01, 0.05, 0}
	expectedVectorB := []float64{0, 0, 0.05, 0.15}
	vectorA, vectorB := util.GetTfIdfVectors(idfMap, tfMap1, tfMap2)
	if len(vectorA) != len(vectorB) || len(vectorA) != len(expectedVectorA) {
		t.Errorf("Vector A or Vector B not of the expected size.")
		return
	}
	vectorsDontMatch := false
	for i := range vectorA {
		foundExpectedMatch := false
		for j := range expectedVectorA {
			if vectorA[i] == expectedVectorA[j] && vectorB[i] == expectedVectorB[j] {
				foundExpectedMatch = true
				break
			}
		}
		// If at least one pair of vectorA[i], vectorB[i] doesn't match any pair of
		// expectedVectorA[j], expectedVectorB[j] then this pair is wrong.
		if !foundExpectedMatch {
			vectorsDontMatch = true
			break
		}
	}
	if vectorsDontMatch {
		t.Error("Vectors are not as expected.")
		t.Errorf("Got: VectorA=%v, VectorB=%v", vectorA, vectorB)
		t.Errorf("Expected: VectorA=%v, VectorB=%v", expectedVectorA, expectedVectorB)
	}
}
