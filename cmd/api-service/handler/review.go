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

	subPath, subTrail := ShiftPath(r.URL.Path)
	subSubPath, _ := ShiftPath(subTrail)
	switch {
	case r.Method == http.MethodGet && subPath == "":
		reviewAPI.List(w, r)
	case r.Method == http.MethodGet && subPath == "unrated" && subSubPath == "":
		reviewAPI.ListUnrated(w, r)
	case r.Method == http.MethodGet && subPath == "unrated" && subSubPath == "next":
		reviewAPI.NextUnrated(w, r)
	case r.Method == http.MethodGet && subPath == "no-titles" && subSubPath == "":
		reviewAPI.ListNoTitles(w, r)
	case r.Method == http.MethodGet && subPath == "no-titles" && subSubPath == "next":
		reviewAPI.NextNoTitles(w, r)
	case r.Method == http.MethodGet && subPath != "":
		reviewAPI.Get(w, r, subPath)
	case r.Method == http.MethodPut && subPath != "":
		reviewAPI.Store(w, r, subPath)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /review", r.Method, subPath), logger)
	}
}

func (reviewAPI *ReviewAPI) Get(w http.ResponseWriter, r *http.Request, id string) {
	logger := reviewAPI.logger.With("method", "get")

	review, err := reviewAPI.repo.FindOne(id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get review", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(review); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode review", err, logger)
		return
	}
}

func (reviewAPI *ReviewAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "list")

	reviews, err := reviewAPI.repo.FindAll()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get reviews", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode reviews", err, logger)
		return
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

func (reviewAPI *ReviewAPI) NextUnrated(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "nextUnrated")

	review, err := reviewAPI.repo.FindNextUnrated()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get review", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(review); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode review", err, logger)
		return
	}
}

func (reviewAPI *ReviewAPI) ListNoTitles(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "listNoTitles")

	reviews, err := reviewAPI.repo.FindNoTitles()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get reviews", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode reviews", err, logger)
		return
	}
}

func (reviewAPI *ReviewAPI) NextNoTitles(w http.ResponseWriter, r *http.Request) {
	logger := reviewAPI.logger.With("method", "nextNoTitles")

	review, err := reviewAPI.repo.FindNextNoTitles()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get review", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(review); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode review", err, logger)
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
