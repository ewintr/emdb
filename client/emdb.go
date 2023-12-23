package client

import (
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
