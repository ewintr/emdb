package moviestore

import (
	"encoding/json"
)

const (
	ReviewSourceIMDB = "imdb"

	MentionsSeparator = "|"
)

type ReviewSource string

type Titles struct {
	Movies  []string `json:"movies"`
	TVShows []string `json:"tvShows"`
	Games   []string `json:"games"`
	Books   []string `json:"books"`
}

type Review struct {
	ID          string
	MovieID     string
	Source      ReviewSource
	URL         string
	Review      string
	MovieRating int
	Quality     int
	Titles      Titles
}

type ReviewRepository struct {
	db *SQLite
}

func NewReviewRepository(db *SQLite) *ReviewRepository {
	return &ReviewRepository{
		db: db,
	}
}

func (rr *ReviewRepository) Store(r Review) error {
	titles, err := json.Marshal(r.Titles)
	if err != nil {
		return err
	}
	if _, err := rr.db.Exec(`REPLACE INTO review (id, movie_id, source, url, review, movie_rating, quality, mentioned_titles) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.MovieID, r.Source, r.URL, r.Review, r.MovieRating, r.Quality, titles); err != nil {
		return err
	}

	return nil
}

func (rr *ReviewRepository) FindOne(id string) (Review, error) {
	row := rr.db.QueryRow(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles FROM review WHERE id=?`, id)
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
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles FROM review WHERE movie_id=?`, movieID)
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
	row := rr.db.QueryRow(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles FROM review WHERE quality=0 LIMIT 1`)
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
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles FROM review WHERE quality=0`)
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

func (rr *ReviewRepository) FindNoTitles() ([]Review, error) {
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles FROM review WHERE mentioned_titles='{}'`)
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
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentioned_titles FROM review`)
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
	if _, err := rr.db.Exec(`DELETE FROM review WHERE movie_id=?`, id); err != nil {
		return err
	}

	return nil
}
