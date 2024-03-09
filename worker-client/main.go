package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/job"
	"code.ewintr.nl/emdb/storage"
	"code.ewintr.nl/emdb/worker-client/worker"
)

const (
	mentionsTemplate = `The following text is a user comment about the movie {{.title}}. In it, the user may have referenced other movie titles. List them if you see any.

----
{{.review}}
---- 

If you found any movie titles other than {{.title}}, list them below in a JSON array. If there are other titles, like TV shows, books or games, ignore them. The format is as follows:

["movie title 1", "movie title 2"]

Just answer with the JSON and nothing else. If you don't see any other movie titles, just answer with an empty array.`
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dbHost := os.Getenv("EMDB_DB_HOST")
	dbName := os.Getenv("EMDB_DB_NAME")
	dbUser := os.Getenv("EMDB_DB_USER")
	dbPassword := os.Getenv("EMDB_DB_PASSWORD")
	pgConnStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	dbPostgres, err := storage.NewPostgres(pgConnStr)
	if err != nil {
		fmt.Printf("could not create new postgres repo: %s", err.Error())
		os.Exit(1)
	}
	movieRepo := storage.NewMovieRepository(dbPostgres)
	reviewRepo := storage.NewReviewRepository(dbPostgres)
	jobQueue := job.NewJobQueue(dbPostgres, logger)

	w := worker.NewWorker(jobQueue, movieRepo, reviewRepo, client.NewIMDB(), logger)

	go w.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
