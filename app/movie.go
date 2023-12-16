package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type Movie struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Year      int    `json:"year"`
	IMDBID    string `json:"imdb_id"`
	WatchedOn string `json:"watched_on"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}

type MovieAPI struct {
	repo   *SQLite
	logger *slog.Logger
}

func NewMovieAPI(repo *SQLite, logger *slog.Logger) *MovieAPI {
	return &MovieAPI{
		repo:   repo,
		logger: logger.With("api", "movie"),
	}
}

func (api *MovieAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := api.logger.With("method", "serveHTTP")

	movieID, _ := ShiftPath(r.URL.Path)
	switch {
	case r.Method == http.MethodGet && movieID != "":
		api.Read(w, r, movieID)
	case r.Method == http.MethodGet && movieID == "":
		api.List(w, r)
	case r.Method == http.MethodPost:
		api.Create(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", fmt.Errorf("method %q with subpath %q was not registered in /movie", r.Method, movieID), logger)
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

func (api *MovieAPI) Create(w http.ResponseWriter, r *http.Request) {
	logger := api.logger.With("method", "create")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, "could not read body", err, logger)
		return
	}
	defer r.Body.Close()

	var movie *Movie
	if err := json.Unmarshal(body, &movie); err != nil {
		Error(w, http.StatusBadRequest, "could not unmarshal request body", err, logger)
		return
	}
	movie.ID = uuid.New().String()

	if err := api.repo.StoreMovie(movie); err != nil {
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
