package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"recommender/models"
	"reflect"
	"strconv"
	"strings"
)

var dataTypes = map[string]reflect.Type{
	/*
	Users map: {key=userID, value={User}}
	Movies map: {key=movieID, value={Movie}}
	Tags map: {key=movieId, value=[{Tag}]}
	*/
	"User":  reflect.TypeOf(map[int]models.User{}),
	"Movie": reflect.TypeOf(map[int]models.Movie{}),
	"Tag":   reflect.TypeOf(map[int][]models.Tag{}),
}

func LoadData(dataField interface{}, filePath string, dataType string, maxRows ...int) {
	// Check if a certain number of rows was requested to be read
	rowsToRead := -1
	if len(maxRows) > 0 {
		rowsToRead = maxRows[0]
	}

	if err := validateDataType(dataType, dataField); err != nil {
		log.Fatalf("Error: %v", err)
		return
	}

	var data interface{}
	var err error
	switch dataType {
	case "User":
		data, err = readRatings(filePath, rowsToRead)
	case "Movie":
		data, err = readMovies(filePath, rowsToRead)
	case "Tag":
		data, err = readTags(filePath, rowsToRead)
	default:
		log.Fatalf("Error: Unsupported data type: %s", dataType)
		return
	}

	if err != nil {
		log.Fatalf("Error loading data from %s: %v", dataType, err)
		return
	}

	reflect.ValueOf(dataField).Elem().Set(reflect.ValueOf(data))
}

func validateDataType(dataType string, dataField interface{}) error {
	fieldType := reflect.TypeOf(dataField).Elem()
	expectedType, found := dataTypes[dataType]
	if !found {
		return errors.New("Unsupported data type: " + dataType)
	}
	if fieldType != expectedType {
		return fmt.Errorf("Invalid data type %v for %s", fieldType, dataType)
	}
	return nil
}

func readRatings(filePath string, maxRows int) (map[int]models.User, error) {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	users := make(map[int]models.User)

	var rowCount int
	for {
		if maxRows > 0 && rowCount >= maxRows {
			break
		}
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		rowCount++

		userID, _ := strconv.Atoi(record[0])
		movieID, _ := strconv.Atoi(record[1])
		rating, _ := strconv.ParseFloat(record[2], 32)

		user, exists := users[userID]
		if !exists {
			user = models.User{}
		}
		user.UserRatings = append(user.UserRatings, models.UserRating{MovieID: movieID, Rating: float32(rating)})
		users[userID] = user
	}
	return users, nil
}

func readMovies(filePath string, maxRows int) (map[int]models.Movie, error) {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	movies := make(map[int]models.Movie)

	var rowCount int
	for {
		if maxRows > 0 && rowCount >= maxRows {
			break
		}
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		movieID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, errors.New("Invalid movieID")
		}

		movies[movieID] = models.Movie{
			Title: record[1],
		}
	}

	return movies, nil
}

func readTags(filePath string, maxRows int) (map[int][]models.Tag, error) {
	file, reader, _, err := openCSVFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tags := make(map[int][]models.Tag)

	var rowCount int
	for {
		if maxRows > 0 && rowCount >= maxRows {
			break
		}
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if len(record) != 4 {
			return nil, errors.New("Mismatch between the number of fields in the CSV record and the Tag model")
		}

		userID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, errors.New("Invalid userID")
		}

		movieID, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, errors.New("Invalid movieID")
		}

		tag := record[2]

		tagItem := models.Tag{
			UserID: userID,
			Tag:    tag,
		}

		// Use the movieID as the key for the tags map
		tags[movieID] = append(tags[movieID], tagItem)

		maxRows--
		if maxRows == 0 {
			break
		}
	}

	return tags, nil
}

func openCSVFile(filePath string) (*os.File, *csv.Reader, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, err
	}

	reader := csv.NewReader(file)
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

func setField(field reflect.Value, value string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, _ := strconv.Atoi(value)
		field.SetInt(int64(intValue))
	case reflect.Float32, reflect.Float64:
		floatValue, _ := strconv.ParseFloat(value, 64)
		field.SetFloat(floatValue)
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
