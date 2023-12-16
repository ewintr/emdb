package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"ewintr.nl/emdb/app"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port, err := strconv.Atoi(getParam("API_PORT", "8080"))
	if err != nil {
		fmt.Printf("invalid port: %s", err.Error())
		os.Exit(1)
	}
	apiKey := getParam("API_KEY", "hoi")
	repo, err := app.NewSQLite(getParam("DB_PATH", "test.db"))
	if err != nil {
		fmt.Printf("could not create new sqlite repo: %s", err.Error())
		os.Exit(1)
	}

	apis := app.APIIndex{
		"movie": app.NewMovieAPI(repo, logger),
	}

	go http.ListenAndServe(fmt.Sprintf(":%d", port), app.NewServer(apiKey, apis, logger))

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
