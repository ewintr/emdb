package storage

import (
	"encoding/json"
)

const (
	ReviewSourceIMDB = "imdb"

	MentionsSeparator = "|"
)

type ReviewSource string

type Review struct {
	ID          string
	MovieID     string
	Source      ReviewSource
	URL         string
	Review      string
	MovieRating int
	Quality     int
	Titles      []string
}

type ReviewRepository struct {
	db *Postgres
}

func NewReviewRepository(db *Postgres) *ReviewRepository {
	return &ReviewRepository{
		db: db,
	}
}

func (rr *ReviewRepository) Store(r Review) error {
	titles, err := json.Marshal(r.Titles)
	if err != nil {
		return err
	}
	if _, err := rr.db.Exec(`INSERT INTO review (id, movie_id, source, url, review, movie_rating, quality, mentioned_titles) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
ON CONFLICT (id) DO UPDATE SET movie_id = EXCLUDED.movie_id, source = EXCLUDED.source, url = EXCLUDED.url, 
review = EXCLUDED.review, movie_rating = EXCLUDED.movie_rating, quality = EXCLUDED.quality, 
mentioned_titles = EXCLUDED.mentioned_titles;`,
		r.ID, r.MovieID, r.Source, r.URL, r.Review, r.MovieRating, r.Quality, titles); err != nil {
		return err
	}

	return nil
}

func (rr *ReviewRepository) FindOne(id string) (Review, error) {
	row := rr.db.QueryRow(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review 
WHERE id=$1`, id)
	if row.Err() != nil {
		return Review{}, row.Err()
	}

	r := Review{}
	var titles string
	if err := row.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
		return Review{}, err
	}
	if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
		return Review{}, err
	}

	return r, nil
}

func (rr *ReviewRepository) FindByMovieID(movieID string) ([]Review, error) {
	rows, err := rr.db.Query(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review 
WHERE movie_id=$1`, movieID)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var titles string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
			return []Review{}, err
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}

func (rr *ReviewRepository) FindNextUnrated() (Review, error) {
	row := rr.db.QueryRow(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review 
WHERE quality=0 
LIMIT 1`)
	if row.Err() != nil {
		return Review{}, row.Err()
	}

	r := Review{}
	var titles string
	if err := row.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
		return Review{}, err
	}
	if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
		return Review{}, err
	}

	return r, nil
}

func (rr *ReviewRepository) FindUnrated() ([]Review, error) {
	rows, err := rr.db.Query(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review 
WHERE quality=0`)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var titles string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
			return []Review{}, err
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}

func (rr *ReviewRepository) FindNextNoTitles() (Review, error) {
	row := rr.db.QueryRow(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review 
WHERE mentioned_titles='{}' 
LIMIT 1`)
	if row.Err() != nil {
		return Review{}, row.Err()
	}

	r := Review{}
	var titles string
	if err := row.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
		return Review{}, err
	}
	if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
		return Review{}, err
	}

	return r, nil
}

func (rr *ReviewRepository) FindNoTitles() ([]Review, error) {
	rows, err := rr.db.Query(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review 
WHERE mentioned_titles='{}'`)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var titles string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
			return []Review{}, err
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}

func (rr *ReviewRepository) FindAll() ([]Review, error) {
	rows, err := rr.db.Query(`
SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles 
FROM review`)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var titles string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &titles); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(titles), &r.Titles); err != nil {
			return []Review{}, err
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}

func (rr *ReviewRepository) DeleteByMovieID(id string) error {
	if _, err := rr.db.Exec(`DELETE FROM review WHERE movie_id=$1`, id); err != nil {
		return err
	}

	return nil
}
