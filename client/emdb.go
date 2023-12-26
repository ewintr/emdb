package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ewintr.nl/emdb/model"
)

type EMDB struct {
	baseURL string
	apiKey  string
	c       *http.Client
}

func NewEMDB(baseURL string, apiKey string) *EMDB {
	return &EMDB{
		baseURL: baseURL,
		apiKey:  apiKey,
		c:       &http.Client{},
	}
}

func (e *EMDB) GetMovies() ([]model.Movie, error) {
	//var movies []model.Movie
	//for i := 0; i < 5; i++ {
	//	movies = append(movies, model.Movie{
	//		ID:     fmt.Sprintf("id-%d", i),
	//		TMDBID: int64(i),
	//		IMDBID: fmt.Sprintf("tt%07d", i),
	//		Title:  fmt.Sprintf("Movie %d", i),
	//	})
	//}
	//return movies, nil

	url := fmt.Sprintf("%s/movie", e.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var movies []model.Movie
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, err
	}

	return movies, nil
}

func (e *EMDB) CreateMovie(movie model.Movie) (model.Movie, error) {
	body, err := json.Marshal(movie)
	if err != nil {
		return model.Movie{}, err
	}

	url := fmt.Sprintf("%s/movie", e.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return model.Movie{}, err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return model.Movie{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return model.Movie{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	newBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Movie{}, err
	}
	defer resp.Body.Close()

	var newMovie model.Movie
	if err := json.Unmarshal(newBody, &newMovie); err != nil {
		return model.Movie{}, err
	}

	return newMovie, nil
}

func (e *EMDB) UpdateMovie(movie model.Movie) (model.Movie, error) {
	body, err := json.Marshal(movie)
	if err != nil {
		return model.Movie{}, err
	}

	url := fmt.Sprintf("%s/movie/%s", e.baseURL, movie.ID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return model.Movie{}, err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return model.Movie{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return model.Movie{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	newBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Movie{}, err
	}
	defer resp.Body.Close()

	var newMovie model.Movie
	if err := json.Unmarshal(newBody, &newMovie); err != nil {
		return model.Movie{}, err
	}

	return newMovie, nil
}
