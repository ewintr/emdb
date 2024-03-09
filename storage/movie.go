package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Movie struct {
	ID           string   `json:"id"`
	TMDBID       int64    `json:"tmdbID"`
	IMDBID       string   `json:"imdbID"`
	Title        string   `json:"title"`
	EnglishTitle string   `json:"englishTitle"`
	Year         int      `json:"year"`
	Directors    []string `json:"directors"`
	WatchedOn    string   `json:"watchedOn"`
	Rating       int      `json:"rating"`
	Summary      string   `json:"summary"`
	Comment      string   `json:"comment"`
}

type MovieRepository struct {
	db *Postgres
}

func NewMovieRepository(db *Postgres) *MovieRepository {
	return &MovieRepository{
		db: db,
	}
}

func (mr *MovieRepository) Store(m Movie) error {
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
		return fmt.Errorf("%w: %v", ErrPostgresqlFailure, err)
	}

	return nil
}

func (mr *MovieRepository) Delete(id string) error {
	if _, err := mr.db.Exec(`DELETE FROM movie WHERE id=$1`, id); err != nil {
		return fmt.Errorf("%w: %v", ErrPostgresqlFailure, err)
	}

	return nil
}

func (mr *MovieRepository) FindOne(id string) (Movie, error) {
	row := mr.db.QueryRow(`
SELECT id, tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment
FROM movie
WHERE id=$1`, id)
	if row.Err() != nil {
		return Movie{}, row.Err()
	}

	m := Movie{
		ID: id,
	}
	var directors string
	if err := row.Scan(&m.ID, &m.TMDBID, &m.IMDBID, &m.Title, &m.EnglishTitle, &m.Year, &directors, &m.Summary, &m.WatchedOn, &m.Rating, &m.Comment); err != nil {
		return Movie{}, fmt.Errorf("%w: %w", ErrPostgresqlFailure, err)
	}
	m.Directors = strings.Split(directors, ",")

	return m, nil
}

func (mr *MovieRepository) FindAll() ([]Movie, error) {
	rows, err := mr.db.Query(`
SELECT id, tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment
FROM movie`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPostgresqlFailure, err)
	}

	movies := make([]Movie, 0)
	defer rows.Close()
	for rows.Next() {
		m := Movie{}
		var directors string
		if err := rows.Scan(&m.ID, &m.TMDBID, &m.IMDBID, &m.Title, &m.EnglishTitle, &m.Year, &directors, &m.Summary, &m.WatchedOn, &m.Rating, &m.Comment); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPostgresqlFailure, err)
		}
		m.Directors = strings.Split(directors, ",")
		movies = append(movies, m)
	}

	return movies, nil
}
