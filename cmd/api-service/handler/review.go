package handler

import (
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
	//case r.Method == http.MethodGet && subPath != "":
	//	reviewAPI.Read(w, r, subPath)
	//case r.Method == http.MethodPut && subPath != "":
	//	reviewAPI.Store(w, r, subPath)
	//case r.Method == http.MethodPost && subPath == "":
	//	reviewAPI.Store(w, r, "")
	//case r.Method == http.MethodDelete && subPath != "":
	//	reviewAPI.Delete(w, r, subPath)
	//case r.Method == http.MethodGet && subPath == "":
	//	reviewAPI.List(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /review", r.Method, subPath), logger)
	}
}
