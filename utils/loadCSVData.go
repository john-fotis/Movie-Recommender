package util

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"recommender/helpers"
	model "recommender/models"
	"reflect"
	"strconv"
	"strings"
)

func LoadCSVData(dataField interface{}, filePath string, maxRows ...int) {
	// Check if a certain number of rows was requested to be read
	rowsToRead := -1
	if len(maxRows) > 0 {
		rowsToRead = maxRows[0]
	}
	var data interface{}
	fieldType := reflect.TypeOf(dataField).Elem()
	for dataType, typeVal := range DataTypes {
		if fieldType == typeVal {
			switch dataType {
			case "UserRatings":
				data = loadUserRatings(filePath, rowsToRead)
			case "MovieTitle":
				data = loadMovieTitles(filePath, rowsToRead)
			case "MovieRatings":
				data = loadMovies(filePath, rowsToRead)
			case "MovieTags":
				data = loadTags(filePath, rowsToRead)
			default:
				log.Fatalf("Unsupported data type: %v", fieldType)
				return
			}
		}
	}
	if data != nil {
		reflect.ValueOf(dataField).Elem().Set(reflect.ValueOf(data))
	}
}

func loadUserRatings(filePath string, maxRows int) map[int]model.User {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to open file %s", filePath)))
		return nil
	}
	defer file.Close()
	var rowCount int
	users := map[int]model.User{}
	for maxRows == -1 || rowCount >= maxRows {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(errors.New("Unexpected end of file"))
			return nil
		}
		rowCount++
		userID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(errors.New("Invalid userID"))
			return nil
		}
		movieID, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatal(errors.New("Invalid movieID"))
			return nil
		}
		rating, err := strconv.ParseFloat(record[2], 32)
		if err != nil {
			log.Fatal(errors.New("Invalid rating"))
			return nil
		}
		user, exists := users[userID]
		if !exists {
			user = model.User{
				MovieRatings: make(map[int]float32),
			}
		}
		user.MovieRatings[movieID] = float32(rating)
		users[userID] = user
	}
	return users
}

func loadMovieTitles(filePath string, maxRows int) map[int]model.MovieTitle {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		log.Fatal(errors.New("Failed to open file"))
		return nil
	}
	defer file.Close()
	movieTitles := map[int]model.MovieTitle{}
	var rowCount int
	for maxRows == -1 || rowCount >= maxRows {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(errors.New("Unexpected end of file"))
			return nil
		}
		rowCount++
		movieID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(errors.New("Invalid movieID"))
			return nil
		}
		movieTitles[movieID] = model.MovieTitle{
			Title: strings.Trim(record[1], "\""),
		}
	}
	return movieTitles
}

func loadMovies(filePath string, maxRows int) map[int]model.Movie {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		log.Fatal(errors.New("Failed to open file"))
		return nil
	}
	defer file.Close()
	movies := map[int]model.Movie{}
	var rowCount int
	for maxRows == -1 || rowCount >= maxRows {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(errors.New("Unexpected end of file"))
			return nil
		}
		rowCount++
		userID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(errors.New("Invalid userID"))
			return nil
		}
		movieID, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatal(errors.New("Invalid movieID"))
			return nil
		}
		rating, err := strconv.ParseFloat(record[2], 32)
		if err != nil {
			log.Fatal(errors.New("Invalid rating"))
			return nil
		}
		movie, exists := movies[movieID]
		if !exists {
			movie = model.Movie{
				UserRatings: make(map[int]float32),
			}
		}
		movie.UserRatings[userID] = float32(rating)
		movies[movieID] = movie
	}
	return movies
}

func loadTags(filePath string, maxRows int) map[int]model.MovieTags {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		log.Fatal(errors.New("Failed to open file"))
		return nil
	}
	defer file.Close()
	tags := map[int]model.MovieTags{}
	var rowCount int
	for maxRows == -1 || rowCount >= maxRows {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(errors.New("Unexpected end of file."))
			return nil
		}
		userID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(errors.New("Invalid userID"))
			return nil
		}
		movieID, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatal(errors.New("Invalid movieID"))
			return nil
		}
		tag, exists := tags[movieID]
		if !exists {
			tag = model.MovieTags{
				UserTags: make(map[int]model.UserTags),
			}
		}
		if userTag, userExists := tag.UserTags[userID]; userExists {
			userTag.Tags = append(userTag.Tags, strings.Join(helpers.ExtractTokensFromStr(record[2]), " "))
			tag.UserTags[userID] = userTag
		} else {
			userTag := model.UserTags{
				Tags: []string{strings.Join(helpers.ExtractTokensFromStr(record[2]), " ")},
			}
			tag.UserTags[userID] = userTag
		}
		// Use the movieID as the key for the tags map
		tags[movieID] = tag
	}
	return tags
}

func openCSVFile(filePath string) (*os.File, *csv.Reader, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, err
	}
	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ','
	// Consume header row
	header, err := reader.Read()
	if err != nil {
		file.Close()
		return nil, nil, nil, err
	}
	// Remove Byte Order Mark if detected
	if strings.Contains(header[0], "\ufeff") {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}

	return file, reader, header, nil
}

func getFieldNames(recordType interface{}) []string {
	elemType := reflect.TypeOf(recordType)
	fieldNames := make([]string, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		fieldNames[i] = elemType.Field(i).Name
	}
	return fieldNames
}

func getFieldIndexes(header []string, fieldNames []string) []int {
	fieldIndexes := make([]int, len(fieldNames))
	for i, fieldName := range fieldNames {
		fieldIndexes[i] = getColumnIndex(header, fieldName)
		if fieldIndexes[i] == -1 {
			return nil
		}
	}
	return fieldIndexes
}

func setField(fieldType reflect.Value, value string) {
	switch fieldType.Kind() {
	case reflect.String:
		fieldType.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, _ := strconv.Atoi(value)
		fieldType.SetInt(int64(intValue))
	case reflect.Float32, reflect.Float64:
		floatValue, _ := strconv.ParseFloat(value, 64)
		fieldType.SetFloat(floatValue)
	}
}

func getColumnIndex(header []string, columnName string) int {
	columnNameLower := strings.ToLower(strings.TrimSpace(columnName))
	for index, column := range header {
		colLower := strings.ToLower(strings.TrimSpace(column))
		if colLower == columnNameLower {
			return index
		}
	}
	return -1
}
