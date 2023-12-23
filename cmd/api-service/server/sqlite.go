package server

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ewintr.nl/emdb/model"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type sqliteMigration string

var sqliteMigrations = []sqliteMigration{
	`CREATE TABLE movie (
		"id" TEXT UNIQUE NOT NULL, 
		"imdb_id" TEXT NOT NULL DEFAULT "",
		"title" TEXT NOT NULL DEFAULT "",
		"english_title" TEXT NOT NULL DEFAULT "",
		"year" INTEGER NOT NULL DEFAULT 0, 
		"directors" TEXT NOT NULL DEFAULT "",
		"watched_on" TEXT NOT NULL DEFAULT "", 
		"rating" INTEGER NOT NULL DEFAULT 0, 
		"comment" TEXT NOT NULL DEFAULT ""
	)`,
	`CREATE TABLE system ("latest_sync" INTEGER)`,
	`INSERT INTO system (latest_sync) VALUES (0)`,
	`ALTER TABLE movie ADD COLUMN tmdb_id INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE movie ADD COLUMN summary TEXT NOT NULL DEFAULT ""`,
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

func (s *SQLite) Store(m *model.Movie) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}
	defer tx.Rollback()

	directors := strings.Join(m.Directors, ",")
	if _, err := s.db.Exec(`REPLACE INTO movie (id, tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.IMDBID, m.IMDBID, m.Title, m.EnglishTitle, m.Year, directors, m.Summary, m.WatchedOn, m.Rating, m.Comment); err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	return nil
}

func (s *SQLite) Delete(id string) error {
	if _, err := s.db.Exec(`DELETE FROM movie WHERE id=?`, id); err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	return nil
}

func (s *SQLite) FindOne(id string) (*model.Movie, error) {
	row := s.db.QueryRow(`
SELECT tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment
FROM movie
WHERE id=?`, id)
	if row.Err() != nil {
		return nil, row.Err()
	}

	m := &model.Movie{
		ID: id,
	}
	var directors string
	if err := row.Scan(&m.IMDBID, &m.IMDBID, &m.Title, &m.EnglishTitle, &m.Year, &directors, &m.Summary, &m.WatchedOn, &m.Rating, &m.Comment); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}
	m.Directors = strings.Split(directors, ",")

	return m, nil
}

func (s *SQLite) FindAll() ([]*model.Movie, error) {
	rows, err := s.db.Query(`
SELECT tmdb_id, imdb_id, title, english_title, year, directors, summary, watched_on, rating, comment
FROM movie`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	movies := make([]*model.Movie, 0)
	defer rows.Close()
	for rows.Next() {
		m := &model.Movie{}
		var directors string
		if err := rows.Scan(&m.TMDBID, &m.IMDBID, &m.Title, &m.EnglishTitle, &m.Year, &directors, &m.Summary, &m.WatchedOn, &m.Rating, &m.Comment); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}
		m.Directors = strings.Split(directors, ",")
		movies = append(movies, m)
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
