package main

import (
	"fmt"
	"log"
	"recommender/config"
	model "recommender/models"
	util "recommender/utils"
	"runtime"
)

// Data represents the data structures for different CSV files.
type Data struct {
	Movies map[int]model.Movie
	Users  map[int]model.User
	Tags   map[int][]model.Tag
}

func main() {
	var m runtime.MemStats

	cfg, err := config.InitServer()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
		return
	}
	
	data := Data{}
	util.LoadData(&data.Movies, cfg.DataDir+"/movies.csv", "Movie")
	fmt.Println(len(data.Movies))
	fmt.Println(data.Movies[1])
	util.LoadData(&data.Users, cfg.DataDir+"/ratings.csv", "User")
	fmt.Println(len(data.Users))
	fmt.Println(data.Users[19705])
	// util.LoadData(&data.Tags, cfg.DataDir+"/tags.csv", "Tag")
	// total := 0
	// for _, tags := range data.Tags {
	// 	total += len(tags)
	// }
	// fmt.Println(total)
	// fmt.Println(data.Tags[1])

	runtime.ReadMemStats(&m)
	fmt.Printf("HeapAlloc: %d MiB\n", m.HeapAlloc/1024/1024)
}
