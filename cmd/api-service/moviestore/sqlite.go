package moviestore

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	`BEGIN TRANSACTION;
		CREATE TABLE movie_new (
			"id" TEXT UNIQUE NOT NULL,
			"imdb_id" TEXT UNIQUE NOT NULL DEFAULT "",
			"tmdb_id" INTEGER UNIQUE NOT NULL DEFAULT 0,
			"title" TEXT NOT NULL DEFAULT "",
			"english_title" TEXT NOT NULL DEFAULT "",
			"year" INTEGER NOT NULL DEFAULT 0,
			"directors" TEXT NOT NULL DEFAULT "",
			"summary" TEXT NOT NULL DEFAULT "",
			"watched_on" TEXT NOT NULL DEFAULT "",
			"rating" INTEGER NOT NULL DEFAULT 0,
			"comment" TEXT NOT NULL DEFAULT ""
		);
		INSERT INTO movie_new (id, imdb_id, tmdb_id, title, english_title, year, directors, summary, watched_on, rating, comment)
		SELECT id, imdb_id, tmdb_id, title, english_title, year, directors, summary, watched_on, rating, comment FROM movie;
		DROP TABLE movie;
		ALTER TABLE movie_new RENAME TO movie;
	COMMIT`,
	`CREATE TABLE review (
		"id" TEXT UNIQUE NOT NULL,
		"movie_id" TEXT NOT NULL,
		"source" TEXT NOT NULL DEFAULT "",
		"url" TEXT NOT NULL DEFAULT "",
		"review" TEXT NOT NULL DEFAULT ""
	)`,
	`CREATE TABLE job_queue (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"movie_id" TEXT NOT NULL,
		"action" TEXT NOT NULL DEFAULT "",
		"status" TEXT NOT NULL DEFAULT ""
	)`,
	`PRAGMA journal_mode=WAL`,
	`INSERT INTO job_queue (movie_id, action, status)
		SELECT id, 'fetch-imdb-reviews', 'todo'
		FROM movie`,
	`AlTER TABLE review ADD COLUMN "references" TEXT NOT NULL DEFAULT ""`,
	`ALTER TABLE review ADD COLUMN "quality" INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE review DROP COLUMN "references"`,
	`ALTER TABLE review ADD COLUMN "mentions" TEXT NOT NULL DEFAULT ""`,
	`ALTER TABLE review ADD COLUMN "movie_rating" INTEGER NOT NULL DEFAULT 0`,
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

	_, err = db.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d;", 5*time.Second))

	s := &SQLite{
		db: db,
	}

	if err := s.migrate(sqliteMigrations); err != nil {
		return &SQLite{}, err
	}

	return s, nil
}

func (s *SQLite) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *SQLite) QueryRow(query string, args ...any) *sql.Row {
	return s.db.QueryRow(query, args...)
}

func (s *SQLite) Query(query string, args ...any) (*sql.Rows, error) {
	return s.db.Query(query, args...)
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
