package util

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	model "recommender/models"
	"reflect"
	"sort"
)

var DataTypes = map[string]reflect.Type{
	"UserRatings":  reflect.TypeOf(map[int]model.User{}),
	"MovieTitle":   reflect.TypeOf(map[int]model.MovieTitle{}),
	"MovieRatings": reflect.TypeOf(map[int]model.Movie{}),
	"MovieTags":    reflect.TypeOf(map[int]model.MovieTags{}),
}

/*
Valid datatypes for $dataField:
  - UserRatings map:  map[int]model.User{}}
  - MovieTitles map:  map[int]model.MovieTitle{}}
  - MovieRatings map: map[int]model.Movie{}}
  - MovieTags map:    map[int]model.MovieTags{}}
*/
func LoadData(dataField interface{}, filePath string, maxRecords ...int) {
	rowsToRead := -1
	if len(maxRecords) > 0 {
		rowsToRead = maxRecords[0]
	}
	fieldType := reflect.TypeOf(dataField).Elem()
	var data interface{}
	for dataType, typeVal := range DataTypes {
		if fieldType == typeVal {
			switch dataType {
			case "UserRatings":
				data = loadProcessedData(filePath, rowsToRead, decodeUser)
			case "MovieTitle":
				data = loadProcessedData(filePath, rowsToRead, decodeMovieTitle)
			case "MovieRatings":
				data = loadProcessedData(filePath, rowsToRead, decodeMovie)
			case "MovieTags":
				data = loadProcessedData(filePath, rowsToRead, decodeMovieTags)
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

func loadProcessedData(filePath string, maxRecords int, decodeFunc func(*gob.Decoder) interface{}) interface{} {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to open file %s", filePath)))
		return nil
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	data := decodeFunc(decoder)
	if maxRecords != -1 {
		dataValue := reflect.ValueOf(data)
		limitedData := reflect.MakeMap(reflect.TypeOf(dataValue.Interface()))
		keys := dataValue.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Interface().(int) < keys[j].Interface().(int)
		})
		count := 0
		for _, key := range keys {
			if count >= maxRecords {
				break
			}
			value := dataValue.MapIndex(key)
			limitedData.SetMapIndex(key, value)
			count++
		}
		return limitedData.Interface()
	}
	return data
}

func decodeUser(decoder *gob.Decoder) interface{} {
	var data map[int]model.User
	if err := decoder.Decode(&data); err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to decode user data")))
		return nil
	}
	return data
}

func decodeMovieTitle(decoder *gob.Decoder) interface{} {
	var data map[int]model.MovieTitle
	if err := decoder.Decode(&data); err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to decode title data")))
		return nil
	}
	return data
}

func decodeMovie(decoder *gob.Decoder) interface{} {
	var data map[int]model.Movie
	if err := decoder.Decode(&data); err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to decode movie data")))
		return nil
	}
	return data
}

func decodeMovieTags(decoder *gob.Decoder) interface{} {
	var data map[int]model.MovieTags
	if err := decoder.Decode(&data); err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to decode tag data")))
		return nil
	}
	return data
}
