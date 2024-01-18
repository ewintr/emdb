package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ewintr.nl/emdb/cmd/api-service/moviestore"
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

func (e *EMDB) GetMovies() ([]moviestore.Movie, error) {
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

	var movies []moviestore.Movie
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, err
	}

	return movies, nil
}

func (e *EMDB) CreateMovie(m moviestore.Movie) (moviestore.Movie, error) {
	body, err := json.Marshal(m)
	if err != nil {
		return moviestore.Movie{}, err
	}

	url := fmt.Sprintf("%s/movie", e.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return moviestore.Movie{}, err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return moviestore.Movie{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return moviestore.Movie{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	newBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return moviestore.Movie{}, err
	}
	defer resp.Body.Close()

	var newMovie moviestore.Movie
	if err := json.Unmarshal(newBody, &newMovie); err != nil {
		return moviestore.Movie{}, err
	}

	return newMovie, nil
}

func (e *EMDB) UpdateMovie(m moviestore.Movie) (moviestore.Movie, error) {
	body, err := json.Marshal(m)
	if err != nil {
		return moviestore.Movie{}, err
	}

	url := fmt.Sprintf("%s/movie/%s", e.baseURL, m.ID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return moviestore.Movie{}, err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return moviestore.Movie{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return moviestore.Movie{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	newBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return moviestore.Movie{}, err
	}
	defer resp.Body.Close()

	var newMovie moviestore.Movie
	if err := json.Unmarshal(newBody, &newMovie); err != nil {
		return moviestore.Movie{}, err
	}

	return newMovie, nil
}

func (e *EMDB) GetReviews(movieID string) ([]moviestore.Review, error) {
	url := fmt.Sprintf("%s/movie/%s/review", e.baseURL, movieID)
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

	var reviews []moviestore.Review
	if err := json.Unmarshal(body, &reviews); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (e *EMDB) GetNextUnratedReview() (moviestore.Review, error) {
	url := fmt.Sprintf("%s/review/unrated/next", e.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return moviestore.Review{}, err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return moviestore.Review{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return moviestore.Review{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var review moviestore.Review
	if err := json.Unmarshal(body, &review); err != nil {
		return moviestore.Review{}, err
	}

	return review, nil
}

func (e *EMDB) UpdateReview(review moviestore.Review) error {
	body, err := json.Marshal(review)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/review/%s", e.baseURL, review.ID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (e *EMDB) CreateJob(movieID, action string) error {
	j := struct {
		MovieID string
		Action  string
	}{
		MovieID: movieID,
		Action:  action,
	}

	body, err := json.Marshal(j)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/job", e.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", e.apiKey)

	resp, err := e.c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
