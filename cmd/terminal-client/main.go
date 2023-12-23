package main

import (
	"fmt"
	"os"

	"ewintr.nl/emdb/cmd/terminal-client/tui"
)

func main() {
	logger := tui.NewLogger()

	p, err := tui.New(tui.Config{
		TMDBAPIKey:  os.Getenv("TMDB_API_KEY"),
		EMDBAPIKey:  os.Getenv("EMDB_API_KEY"),
		EMDBBaseURL: "https://emdb.ewintr.nl",
	}, logger)
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
