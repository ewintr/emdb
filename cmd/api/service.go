package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ewintr.nl/emdb/cmd/api/server"
)

var (
	port   = flag.Int("port", 8080, "port to listen on")
	dbPath = flag.String("dbpath", "test.db", "path to sqlite db")
	apiKey = flag.String("apikey", "hoi", "api key to use")
)

func main() {
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting server", "port", *port, "dbPath", *dbPath)

	repo, err := server.NewSQLite(*dbPath)
	if err != nil {
		fmt.Printf("could not create new sqlite repo: %s", err.Error())
		os.Exit(1)
	}

	apis := server.APIIndex{
		"movie": server.NewMovieAPI(repo, logger),
	}

	go http.ListenAndServe(fmt.Sprintf(":%d", *port), server.NewServer(*apiKey, apis, logger))
	logger.Info("server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logger.Info("server stopped")
}
