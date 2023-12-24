package main

import (
	"fmt"
	"os"

	"ewintr.nl/emdb/client"
	"ewintr.nl/emdb/cmd/terminal-client/tui"
)

func main() {
	tmdb, err := client.NewTMDB(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	emdb := client.NewEMDB("https://emdb.ewintr.nl", os.Getenv("EMDB_API_KEY"))

	p, err := tui.New(emdb, tmdb)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
