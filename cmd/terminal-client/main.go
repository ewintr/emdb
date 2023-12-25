package main

import (
	"fmt"
	"os"

	"ewintr.nl/emdb/client"
	"ewintr.nl/emdb/cmd/terminal-client/tui"
)

func main() {
	logger := tui.NewLogger()
	tmdb, err := client.NewTMDB(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	emdb := client.NewEMDB(os.Getenv("EMDB_BASE_URL"), os.Getenv("EMDB_API_KEY"))

	p, err := tui.New(emdb, tmdb, logger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	logger.SetProgram(p)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
