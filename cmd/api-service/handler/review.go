package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"ewintr.nl/emdb/cmd/api-service/moviestore"
)

type ReviewAPI struct {
	repo   *moviestore.ReviewRepository
	logger *slog.Logger
}

func NewReviewAPI(repo *moviestore.ReviewRepository, logger *slog.Logger) *ReviewAPI {
	return &ReviewAPI{
		repo:   repo,
		logger: logger.With("api", "review"),
	}
}

func (reviewAPI *ReviewAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "serveHTTP")

	subPath, _ := ShiftPath(r.URL.Path)
	switch {
	case r.Method == http.MethodGet && subPath == "":
		reviewAPI.List(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /review", r.Method, subPath), logger)
	}
}

func (reviewAPI *ReviewAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "list")

	movieID := r.Context().Value(MovieKey).(string)
	reviews, err := reviewAPI.repo.FindByMovieID(movieID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get reviews", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode reviews", err, logger)
		return
	}
}
