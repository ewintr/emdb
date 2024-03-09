package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/cmd/api-service/handler"
	"code.ewintr.nl/emdb/cmd/api-service/moviestore"
	job2 "code.ewintr.nl/emdb/job"
	"code.ewintr.nl/emdb/storage"
)

var (
	port   = flag.Int("port", 8085, "port to listen on")
	dbPath = flag.String("dbpath", "test.db", "path to sqlite db")
	apiKey = flag.String("apikey", "hoi", "api key to use")
)

func main() {
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting server", "port", *port, "dbPath", *dbPath)

	db, err := moviestore.NewSQLite(*dbPath)
	if err != nil {
		fmt.Printf("could not create new sqlite repo: %s", err.Error())
		os.Exit(1)
	}

	jobQueue := job2.NewJobQueue(db, logger)
	worker := job2.NewWorker(jobQueue, storage.NewMovieRepository(db), storage.NewReviewRepository(db), client.NewIMDB(), logger)
	go worker.Run()

	apis := handler.APIIndex{
		"job": handler.NewJobAPI(jobQueue, logger),
		"movie": handler.NewMovieAPI(handler.APIIndex{
			"review": handler.NewMovieReviewAPI(storage.NewReviewRepository(db), logger),
		}, storage.NewMovieRepository(db), jobQueue, logger),
		"review": handler.NewReviewAPI(storage.NewReviewRepository(db), logger),
	}

	go http.ListenAndServe(fmt.Sprintf(":%d", *port), handler.NewServer(*apiKey, apis, logger))
	logger.Info("server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logger.Info("server stopped")
}
