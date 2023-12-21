package main

import (
	"fmt"
	"net/http"
	"os"

	"ewintr.nl/emdb/cmd/cli/clients"
)

func main() {
	tdb, err := clients.NewTMDB(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	movies, err := tdb.Search("stark fear")
	for _, m := range movies.Results {
		fmt.Printf("result: %+v\n", m)
	}

}

type EMDBClient struct {
	c *http.Client
}
