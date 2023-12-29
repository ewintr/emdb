package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"ewintr.nl/emdb/cmd/api-service/moviestore"
	"github.com/google/uuid"
)

type MovieAPI struct {
	apis   APIIndex
	repo   *moviestore.MovieRepository
	jq     *moviestore.JobQueue
	logger *slog.Logger
}

func NewMovieAPI(apis APIIndex, repo *moviestore.MovieRepository, jq *moviestore.JobQueue, logger *slog.Logger) *MovieAPI {
	return &MovieAPI{
		apis:   apis,
		repo:   repo,
		jq:     jq,
		logger: logger.With("api", "movie"),
	}
}

func (movieAPI *MovieAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := movieAPI.logger.With("method", "serveHTTP")

	subPath, subTail := ShiftPath(r.URL.Path)
	for aPath, api := range movieAPI.apis {
		if subPath == aPath {
			r.URL.Path = subTail
			r = r.Clone(context.WithValue(r.Context(), MovieKey, subPath))
			api.ServeHTTP(w, r)
			return
		}
	}

	switch {
	case r.Method == http.MethodGet && subPath != "":
		movieAPI.Read(w, r, subPath)
	case r.Method == http.MethodPut && subPath != "":
		movieAPI.Store(w, r, subPath)
	case r.Method == http.MethodPost && subPath == "":
		movieAPI.Store(w, r, "")
	case r.Method == http.MethodDelete && subPath != "":
		movieAPI.Delete(w, r, subPath)
	case r.Method == http.MethodGet && subPath == "":
		movieAPI.List(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /movie", r.Method, subPath), logger)
	}
}

func (movieAPI *MovieAPI) Read(w http.ResponseWriter, r *http.Request, movieID string) {
	logger := movieAPI.logger.With("method", "read")

	m, err := movieAPI.repo.FindOne(movieID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
		return
	case err != nil:
		Error(w, http.StatusInternalServerError, "could not get movie", err, logger)
		return
	}

	resJson, err := json.Marshal(m)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not marshal response", err, logger)
		return
	}

	fmt.Fprint(w, string(resJson))
}

func (movieAPI *MovieAPI) Store(w http.ResponseWriter, r *http.Request, urlID string) {
	logger := movieAPI.logger.With("method", "create")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, "could not read body", err, logger)
		return
	}
	defer r.Body.Close()

	var m moviestore.Movie
	if err := json.Unmarshal(body, &m); err != nil {
		Error(w, http.StatusBadRequest, "could not unmarshal request body", err, logger)
		return
	}

	switch {
	case urlID == "" && m.ID == "":
		m.ID = uuid.New().String()
	case urlID != "" && m.ID == "":
		m.ID = urlID
	case urlID != "" && m.ID != "" && urlID != m.ID:
		Error(w, http.StatusBadRequest, "id in path does not match id in body", err, logger)
		return
	}

	if err := movieAPI.repo.Store(m); err != nil {
		Error(w, http.StatusInternalServerError, "could not store movie", err, logger)
		return
	}

	if err := movieAPI.jq.Add(m.ID, moviestore.ActionFetchIMDBReviews); err != nil {
		Error(w, http.StatusInternalServerError, "could not add job to queue", err, logger)
		return
	}

	resBody, err := json.Marshal(m)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not marshal movie", err, logger)
		return
	}

	fmt.Fprint(w, string(resBody))
}

func (movieAPI *MovieAPI) Delete(w http.ResponseWriter, r *http.Request, urlID string) {
	logger := movieAPI.logger.With("method", "delete")

	err := movieAPI.repo.Delete(urlID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
		return
	case err != nil:
		Error(w, http.StatusInternalServerError, "could not delete movie", err, logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (movieAPI *MovieAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := movieAPI.logger.With("method", "list")

	movies, err := movieAPI.repo.FindAll()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not get movies", err, logger)
		return
	}

	resBody, err := json.Marshal(movies)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not marshal movies", err, logger)
		return
	}

	fmt.Fprint(w, string(resBody))
}
