package moviestore

import "strings"

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
	Mentions    []string
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
	if _, err := rr.db.Exec(`REPLACE INTO review (id, movie_id, source, url, review, movie_rating, quality, mentions) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.MovieID, r.Source, r.URL, r.Review, r.MovieRating, r.Quality, strings.Join(r.Mentions, MentionsSeparator)); err != nil {
		return err
	}

	return nil
}

func (rr *ReviewRepository) FindOne(id string) (Review, error) {
	row := rr.db.QueryRow(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentions FROM review WHERE id=?`, id)
	if row.Err() != nil {
		return Review{}, row.Err()
	}

	r := Review{}
	var mentions string
	if err := row.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &mentions); err != nil {
		return Review{}, err
	}
	r.Mentions = make([]string, 0)
	if mentions != "" {
		r.Mentions = strings.Split(mentions, MentionsSeparator)
	}
	return r, nil
}

func (rr *ReviewRepository) FindByMovieID(movieID string) ([]Review, error) {
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentions FROM review WHERE movie_id=?`, movieID)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var mentions string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &mentions); err != nil {
			return nil, err
		}
		r.Mentions = make([]string, 0)
		if mentions != "" {
			r.Mentions = strings.Split(mentions, MentionsSeparator)
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}

func (rr *ReviewRepository) FindNextUnrated() (Review, error) {
	row := rr.db.QueryRow(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentions FROM review WHERE quality=0 LIMIT 1`)
	if row.Err() != nil {
		return Review{}, row.Err()
	}

	r := Review{}
	var mentions string
	if err := row.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &mentions); err != nil {
		return Review{}, err
	}
	r.Mentions = make([]string, 0)
	if mentions != "" {
		r.Mentions = strings.Split(mentions, MentionsSeparator)
	}

	return r, nil
}

func (rr *ReviewRepository) FindUnrated() ([]Review, error) {
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentions FROM review WHERE quality=0`)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var mentions string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &mentions); err != nil {
			return nil, err
		}
		r.Mentions = make([]string, 0)
		if mentions != "" {
			r.Mentions = strings.Split(mentions, MentionsSeparator)
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}

func (rr *ReviewRepository) FindAll() ([]Review, error) {
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review, movie_rating, quality, mentions FROM review`)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	var mentions string
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review, &r.MovieRating, &r.Quality, &mentions); err != nil {
			return nil, err
		}
		r.Mentions = make([]string, 0)
		if mentions != "" {
			r.Mentions = strings.Split(mentions, MentionsSeparator)
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
