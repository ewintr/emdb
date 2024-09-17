package main

import (
	"fmt"
	"os"

	"go-mod.ewintr.nl/emdb/client"
	"go-mod.ewintr.nl/emdb/desktop-client/backend"
	"go-mod.ewintr.nl/emdb/desktop-client/gui"
	"go-mod.ewintr.nl/emdb/storage"
)

func main() {
	tmdb, err := client.NewTMDB(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

	b := backend.NewBackend(movieRepo, tmdb)
	g := gui.New(b.In(), b.Out())
	g.Run()
}
