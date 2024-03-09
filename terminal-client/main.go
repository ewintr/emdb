package main

import (
	"fmt"
	"log/slog"
	"os"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/job"
	"code.ewintr.nl/emdb/storage"
	"code.ewintr.nl/emdb/terminal-client/tui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tuiLogger := tui.NewLogger()
	tmdb, err := client.NewTMDB(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//emdb := client.NewEMDB(os.Getenv("EMDB_BASE_URL"), os.Getenv("EMDB_API_KEY"))
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

	p, err := tui.New(movieRepo, reviewRepo, jobQueue, tmdb, tuiLogger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tuiLogger.SetProgram(p)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
