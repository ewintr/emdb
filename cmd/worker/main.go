package main

import (
	"fmt"
	"os"

	"ewintr.nl/emdb/client"
)

func main() {
	fmt.Println("worker")

	imdb := client.NewIMDB()
	reviews, err := imdb.GetReviews("tt5540188")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("reviews: %+v", reviews)
}
