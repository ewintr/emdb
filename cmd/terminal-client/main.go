package main

import (
	"fmt"
	"net/http"
	"os"

	"ewintr.nl/emdb/cmd/terminal-client/clients"
	"ewintr.nl/emdb/cmd/terminal-client/tui"
)

func main() {
	tdb, err := clients.NewTMDB(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//
	//movies, err := tdb.Search("stark fear")
	//for _, m := range movies.Results {
	//	fmt.Printf("result: %+v\n", m)
	//}

	p := tui.New(tdb)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

type EMDBClient struct {
	c *http.Client
}
