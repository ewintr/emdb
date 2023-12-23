package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"ewintr.nl/emdb/model"
	"github.com/google/uuid"
)

type MovieAPI struct {
	repo   model.MovieRepository
	logger *slog.Logger
}

func NewMovieAPI(repo model.MovieRepository, logger *slog.Logger) *MovieAPI {
	return &MovieAPI{
		repo:   repo,
		logger: logger.With("api", "movie"),
	}
}

func (api *MovieAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := api.logger.With("method", "serveHTTP")

	subPath, _ := ShiftPath(r.URL.Path)
	switch {
	case r.Method == http.MethodGet && subPath != "":
		api.Read(w, r, subPath)
	case r.Method == http.MethodPut && subPath != "":
		api.Store(w, r, subPath)
	case r.Method == http.MethodPost && subPath == "":
		api.Store(w, r, "")
	case r.Method == http.MethodDelete && subPath != "":
		api.Delete(w, r, subPath)
	case r.Method == http.MethodGet && subPath == "":
		api.List(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /movie", r.Method, subPath), logger)
	}
}

func (api *MovieAPI) Read(w http.ResponseWriter, r *http.Request, movieID string) {
	logger := api.logger.With("method", "read")

	movie, err := api.repo.FindOne(movieID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
		return
	case err != nil:
		Error(w, http.StatusInternalServerError, "could not get movie", err, logger)
		return
	}

	resJson, err := json.Marshal(movie)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not marshal response", err, logger)
		return
	}

	fmt.Fprint(w, string(resJson))
}

func (api *MovieAPI) Store(w http.ResponseWriter, r *http.Request, urlID string) {
	logger := api.logger.With("method", "create")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, "could not read body", err, logger)
		return
	}
	defer r.Body.Close()

	var movie *model.Movie
	if err := json.Unmarshal(body, &movie); err != nil {
		Error(w, http.StatusBadRequest, "could not unmarshal request body", err, logger)
		return
	}

	switch {
	case urlID == "" && movie.ID == "":
		movie.ID = uuid.New().String()
	case urlID != "" && movie.ID == "":
		movie.ID = urlID
	case urlID != "" && movie.ID != "" && urlID != movie.ID:
		Error(w, http.StatusBadRequest, "id in path does not match id in body", err, logger)
		return
	}

	if err := api.repo.Store(movie); err != nil {
		Error(w, http.StatusInternalServerError, "could not store movie", err, logger)
		return
	}

	resBody, err := json.Marshal(movie)
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not marshal movie", err, logger)
		return
	}

	fmt.Fprint(w, string(resBody))
}

func (api *MovieAPI) Delete(w http.ResponseWriter, r *http.Request, urlID string) {
	logger := api.logger.With("method", "delete")

	if err := api.repo.Delete(urlID); err != nil {
		Error(w, http.StatusInternalServerError, "could not delete movie", err, logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *MovieAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := api.logger.With("method", "list")

	movies, err := api.repo.FindAll()
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
