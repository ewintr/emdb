package moviestore

const (
	ReviewSourceIMDB = "imdb"
)

type ReviewSource string

type Review struct {
	ID      string
	MovieID string
	Source  ReviewSource
	URL     string
	Review  string
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
	if _, err := rr.db.Exec(`REPLACE INTO review (id, movie_id, source, url, review) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.MovieID, r.Source, r.URL, r.Review); err != nil {
		return err
	}

	return nil
}

func (rr *ReviewRepository) FindOne(id string) (Review, error) {
	row := rr.db.QueryRow(`SELECT id, movie_id, source, url, review FROM review WHERE id=?`, id)
	if row.Err() != nil {
		return Review{}, row.Err()
	}

	review := Review{}
	if err := row.Scan(&review.ID, &review.MovieID, &review.Source, &review.URL, &review.Review); err != nil {
		return Review{}, err
	}

	return review, nil
}

func (rr *ReviewRepository) FindByMovieID(movieID string) ([]Review, error) {
	rows, err := rr.db.Query(`SELECT id, movie_id, source, url, review FROM review WHERE movie_id=?`, movieID)
	if err != nil {
		return nil, err
	}

	reviews := make([]Review, 0)
	for rows.Next() {
		r := Review{}
		if err := rows.Scan(&r.ID, &r.MovieID, &r.Source, &r.URL, &r.Review); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	rows.Close()

	return reviews, nil
}
