package storage

import (
	"fmt"
	"strings"

	"code.ewintr.nl/emdb/cmd/api-service/moviestore"
	"github.com/google/uuid"
)

type MovieRepositoryPG struct {
	db *Postgres
}

func NewMovieRepositoryPG(db *Postgres) *MovieRepositoryPG {
	return &MovieRepositoryPG{
		db: db,
	}
}

func (mr *MovieRepositoryPG) Store(m moviestore.Movie) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	directors := strings.Join(m.Directors, ",")
	if _, err := mr.db.Exec(`INSERT INTO movie (id, tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE 
SET 
  tmdb_id = EXCLUDED.tmdb_id, 
  imdb_id = EXCLUDED.imdb_id, 
  title = EXCLUDED.title, 
  english_title = EXCLUDED.english_title, 
  year = EXCLUDED.year, 
  directors = EXCLUDED.directors, 
  summary = EXCLUDED.summary, 
  watched_on = EXCLUDED.watched_on, 
  rating = EXCLUDED.rating, 
  comment = EXCLUDED.comment;`,
		m.ID, m.TMDBID, m.IMDBID, m.Title, m.EnglishTitle, m.Year, directors, m.Summary, m.WatchedOn, m.Rating, m.Comment); err != nil {
		return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
	}

	return nil
}

func (mr *MovieRepositoryPG) Delete(id string) error {
	if _, err := mr.db.Exec(`DELETE FROM movie WHERE id=$1`, id); err != nil {
		return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
	}

	return nil
}

func (mr *MovieRepositoryPG) FindOne(id string) (moviestore.Movie, error) {
	row := mr.db.QueryRow(`
SELECT id, tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment
FROM movie
WHERE id=$1`, id)
	if row.Err() != nil {
		return moviestore.Movie{}, row.Err()
	}

	m := moviestore.Movie{
		ID: id,
	}
	var directors string
	if err := row.Scan(&m.ID, &m.TMDBID, &m.IMDBID, &m.Title, &m.EnglishTitle, &m.Year, &directors, &m.Summary, &m.WatchedOn, &m.Rating, &m.Comment); err != nil {
		return moviestore.Movie{}, fmt.Errorf("%w: %w", moviestore.ErrSqliteFailure, err)
	}
	m.Directors = strings.Split(directors, ",")

	return m, nil
}

func (mr *MovieRepositoryPG) FindAll() ([]moviestore.Movie, error) {
	rows, err := mr.db.Query(`
SELECT id, tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment
FROM movie`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
	}

	movies := make([]moviestore.Movie, 0)
	defer rows.Close()
	for rows.Next() {
		m := moviestore.Movie{}
		var directors string
		if err := rows.Scan(&m.ID, &m.TMDBID, &m.IMDBID, &m.Title, &m.EnglishTitle, &m.Year, &directors, &m.Summary, &m.WatchedOn, &m.Rating, &m.Comment); err != nil {
			return nil, fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
		}
		m.Directors = strings.Split(directors, ",")
		movies = append(movies, m)
	}

	return movies, nil
}
