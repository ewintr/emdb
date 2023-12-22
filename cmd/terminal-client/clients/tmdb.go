package clients

import (
	tmdb "github.com/cyruzin/golang-tmdb"
)

type TMDB struct {
	c *tmdb.Client
}

func NewTMDB(apikey string) (*TMDB, error) {
	tmdbClient, err := tmdb.Init(apikey)
	if err != nil {
		return nil, err
	}
	tmdbClient.SetClientAutoRetry()
	tmdbClient.SetAlternateBaseURL()

	return &TMDB{
		c: tmdbClient,
	}, nil
}

func (t TMDB) Search(query string) (*tmdb.SearchMovies, error) {
	return t.c.GetSearchMovies(query, nil)
}
