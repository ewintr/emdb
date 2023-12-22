package clients

import (
	"time"

	"ewintr.nl/emdb/movie"
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

func (t TMDB) Search(query string) ([]movie.Movie, error) {
	return []movie.Movie{
		{Title: "movie1", Year: 2020, Summary: "summary1"},
		{Title: "movie2", Year: 2020, Summary: "summary2"},
		{Title: "movie3", Year: 2020, Summary: "summary3"},
	}, nil

	results, err := t.c.GetSearchMovies(query, nil)
	if err != nil {
		return nil, err
	}

	movies := make([]movie.Movie, len(results.Results))
	for i, result := range results.Results {
		movies[i], err = t.GetMovie(result.ID)
		if err != nil {
			return nil, err
		}
	}

	return movies, nil
}

func (t TMDB) GetMovie(id int64) (movie.Movie, error) {
	result, err := t.c.GetMovieDetails(int(id), nil)
	if err != nil {
		return movie.Movie{}, err
	}

	var year int
	if release, err := time.Parse("2006-01-02", result.ReleaseDate); err == nil {
		year = release.Year()
	}

	return movie.Movie{
		Title:   result.Title,
		TMDBID:  result.ID,
		Year:    year,
		Summary: result.Overview,
	}, nil

}
