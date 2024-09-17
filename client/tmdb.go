package client

import (
	"time"

	"go-mod.ewintr.nl/emdb/storage"
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

func (t TMDB) Search(query string) ([]storage.Movie, error) {
	results, err := t.c.GetSearchMovies(query, nil)
	if err != nil {
		return nil, err
	}

	movies := make([]storage.Movie, len(results.Results))
	for i, result := range results.Results {
		movies[i], err = t.GetMovie(result.ID)
		if err != nil {
			return nil, err
		}
	}

	return movies, nil
}

func (t TMDB) GetMovie(id int64) (storage.Movie, error) {
	result, err := t.c.GetMovieDetails(int(id), map[string]string{
		"append_to_response": "credits",
	})
	if err != nil {
		return storage.Movie{}, err
	}

	var year int
	if release, err := time.Parse("2006-01-02", result.ReleaseDate); err == nil {
		year = release.Year()
	}

	directors := make([]string, 0)
	for crew := range result.Credits.Crew {
		if result.Credits.Crew[crew].Job == "Director" {
			directors = append(directors, result.Credits.Crew[crew].Name)
		}
	}

	return storage.Movie{
		Title:        result.OriginalTitle,
		EnglishTitle: result.Title,
		TMDBID:       result.ID,
		IMDBID:       result.IMDbID,
		Year:         year,
		Directors:    directors,
		Summary:      result.Overview,
	}, nil

}
