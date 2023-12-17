package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ewintr.nl/emdb/app"
)

var (
	port   = flag.Int("port", 8080, "port to listen on")
	dbPath = flag.String("dbpath", "test.db", "path to sqlite db")
	apiKey = flag.String("apikey", "hoi", "api key to use")
)

func main() {
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repo, err := app.NewSQLite(*dbPath)
	if err != nil {
		fmt.Printf("could not create new sqlite repo: %s", err.Error())
		os.Exit(1)
	}

	apis := app.APIIndex{
		"movie": app.NewMovieAPI(repo, logger),
	}

	go http.ListenAndServe(fmt.Sprintf(":%d", *port), app.NewServer(*apiKey, apis, logger))

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
