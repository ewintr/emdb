package app

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type sqliteMigration string

var sqliteMigrations = []sqliteMigration{
	`CREATE TABLE movie ("id" TEXT UNIQUE, "title" TEXT, "year" INTEGER, "imdb_id" TEXT, "watched_on" TEXT, "rating" INTEGER, "comment" TEXT)`,
	`CREATE TABLE system ("latest_sync" INTEGER)`,
	`INSERT INTO system (latest_sync) VALUES (0)`,
}

var (
	ErrInvalidConfiguration     = errors.New("invalid configuration")
	ErrIncompatibleSQLMigration = errors.New("incompatible migration")
	ErrNotEnoughSQLMigrations   = errors.New("already more migrations than wanted")
	ErrSqliteFailure            = errors.New("sqlite returned an error")
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return &SQLite{}, fmt.Errorf("%w: %v", ErrInvalidConfiguration, err)
	}

	s := &SQLite{
		db: db,
	}

	if err := s.migrate(sqliteMigrations); err != nil {
		return &SQLite{}, err
	}

	return s, nil
}

func (s *SQLite) StoreMovie(movie *Movie) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}
	defer tx.Rollback()

	if _, err := s.db.Exec(`REPLACE INTO movie (id, title, year, imdb_id, watched_on, rating, comment) 
	VALUES (?, ?, ?, ?, ?, ?, ?)`,
		movie.ID, movie.Title, movie.Year, movie.IMDBID, movie.WatchedOn, movie.Rating, movie.Comment); err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	return nil
}

func (s *SQLite) FindOne(id string) (*Movie, error) {
	row := s.db.QueryRow(`
SELECT title, year, imdb_id, watched_on, rating, comment
FROM movie
WHERE id=?`, id)
	if row.Err() != nil {
		return nil, row.Err()
	}

	movie := &Movie{
		ID: id,
	}
	if err := row.Scan(&movie.Title, &movie.Year, &movie.IMDBID, &movie.WatchedOn, &movie.Rating, &movie.Comment); err != nil {
		return nil, err
	}

	return movie, nil
}

func (s *SQLite) FindAll() ([]*Movie, error) {
	rows, err := s.db.Query(`
SELECT id, title, year, imdb_id, watched_on, rating, comment
FROM movie`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	movies := make([]*Movie, 0)
	defer rows.Close()
	for rows.Next() {
		var id, title, imdbID, watchedOn, comment string
		var year, rating int
		if err := rows.Scan(&id, &title, &year, &imdbID, &watchedOn, &rating, &comment); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}
		movies = append(movies, &Movie{
			ID:        id,
			Title:     title,
			Year:      year,
			IMDBID:    imdbID,
			WatchedOn: watchedOn,
			Rating:    rating,
			Comment:   comment,
		})
	}

	return movies, nil
}

func (s *SQLite) migrate(wanted []sqliteMigration) error {
	// admin table
	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS migration
("id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "query" TEXT)
`); err != nil {
		return err
	}

	// find existing
	rows, err := s.db.Query(`SELECT query FROM migration ORDER BY id`)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	existing := []sqliteMigration{}
	for rows.Next() {
		var query string
		if err := rows.Scan(&query); err != nil {
			return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}
		existing = append(existing, sqliteMigration(query))
	}
	rows.Close()

	// compare
	missing, err := compareMigrations(wanted, existing)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	// execute missing
	for _, query := range missing {
		if _, err := s.db.Exec(string(query)); err != nil {
			return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}

		// register
		if _, err := s.db.Exec(`
INSERT INTO migration
(query) VALUES (?)
`, query); err != nil {
			return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}
	}

	return nil
}

func compareMigrations(wanted, existing []sqliteMigration) ([]sqliteMigration, error) {
	needed := []sqliteMigration{}
	if len(wanted) < len(existing) {
		return []sqliteMigration{}, ErrNotEnoughSQLMigrations
	}

	for i, want := range wanted {
		switch {
		case i >= len(existing):
			needed = append(needed, want)
		case want == existing[i]:
			// do nothing
		case want != existing[i]:
			return []sqliteMigration{}, fmt.Errorf("%w: %v", ErrIncompatibleSQLMigration, want)
		}
	}

	return needed, nil
}
