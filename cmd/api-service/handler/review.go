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
	case r.Method == http.MethodGet && subPath == "unrated":
		reviewAPI.ListUnrated(w, r)
	case r.Method == http.MethodPut && subPath != "":
		reviewAPI.Store(w, r, subPath)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /review", r.Method, subPath), logger)
	}
}

func (reviewAPI *ReviewAPI) ListUnrated(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "listUnrated")

	reviews, err := reviewAPI.repo.FindUnrated()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get reviews", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode reviews", err, logger)
		return
	}
}

func (reviewAPI *ReviewAPI) Store(w http.ResponseWriter, r *http.Request, id string) {
	logger := reviewAPI.logger.With("method", "store")

	var review moviestore.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		Error(w, http.StatusBadRequest, "could not decode review", err, logger)
		return
	}

	if id != review.ID {
		Error(w, http.StatusBadRequest, "id in path does not match id in body", fmt.Errorf("id in path %q does not match id in body %q", id, review.ID), logger)
		return
	}

	if err := reviewAPI.repo.Store(review); err != nil {
		Error(w, http.StatusInternalServerError, "could not store review", err, logger)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
