package storage

import (
	"database/sql"
	"fmt"

	"code.ewintr.nl/emdb/cmd/api-service/moviestore"
	_ "github.com/lib/pq"
)

type migration string

var migrations = []migration{
	`CREATE TABLE movie (
	"id" TEXT UNIQUE NOT NULL,
	"imdb_id" TEXT NOT NULL DEFAULT '',
	"title" TEXT NOT NULL DEFAULT '',
	"english_title" TEXT NOT NULL DEFAULT '',
	"year" INTEGER NOT NULL DEFAULT 0,
	"directors" TEXT NOT NULL DEFAULT '',
	"watched_on" TEXT NOT NULL DEFAULT '',
	"rating" INTEGER NOT NULL DEFAULT 0,
	"comment" TEXT NOT NULL DEFAULT '',
	"tmdb_id" INTEGER NOT NULL DEFAULT 0,
	"summary" TEXT NOT NULL DEFAULT ''
	);`,
	`CREATE TABLE movie_new (
	"id" TEXT UNIQUE NOT NULL,
	"imdb_id" TEXT UNIQUE NOT NULL DEFAULT '',
	"tmdb_id" INTEGER UNIQUE NOT NULL DEFAULT 0,
	"title" TEXT NOT NULL DEFAULT '',
	"english_title" TEXT NOT NULL DEFAULT '',
	"year" INTEGER NOT NULL DEFAULT 0,
	"directors" TEXT NOT NULL DEFAULT '',
	"summary" TEXT NOT NULL DEFAULT '',
	"watched_on" TEXT NOT NULL DEFAULT '',
	"rating" INTEGER NOT NULL DEFAULT 0,
	"comment" TEXT NOT NULL DEFAULT ''
	);`,
	`CREATE TABLE system ("latest_sync" INTEGER);`,
	`CREATE TABLE review (
	"id" TEXT UNIQUE NOT NULL,
	"movie_id" TEXT NOT NULL,
	"source" TEXT NOT NULL DEFAULT '',
	"url" TEXT NOT NULL DEFAULT '',
	"review" TEXT NOT NULL DEFAULT '',
	"references" TEXT NOT NULL DEFAULT '',
	"quality" INTEGER NOT NULL DEFAULT 0,
	"mentions" TEXT NOT NULL DEFAULT '',
	"movie_rating" INTEGER NOT NULL DEFAULT 0,
	"mentioned_titles" JSONB NOT NULL DEFAULT '[]'
	);`,
	`CREATE TABLE job_queue (
	"id" SERIAL PRIMARY KEY,
	"action_id" TEXT NOT NULL,
	"action" TEXT NOT NULL DEFAULT '',
	"status" TEXT NOT NULL DEFAULT '',
	"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`,
}

type Postgres struct {
	db *sql.DB
}

func NewPostgres(connStr string) (*Postgres, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	pg := &Postgres{
		db: db,
	}

	if err := pg.migrate(migrations); err != nil {
		return &Postgres{}, err
	}

	return pg, nil
}

func (pg *Postgres) migrate(wanted []migration) error {
	// admin table
	if _, err := pg.db.Exec(`
CREATE TABLE IF NOT EXISTS migration
(
    id SERIAL PRIMARY KEY, 
    query TEXT
)`); err != nil {
		return err
	}

	// find existing
	rows, err := pg.db.Query(`SELECT query FROM migration ORDER BY id`)
	if err != nil {
		return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
	}

	existing := []migration{}
	for rows.Next() {
		var query string
		if err := rows.Scan(&query); err != nil {
			return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
		}
		existing = append(existing, migration(query))
	}
	rows.Close()

	// compare
	missing, err := compareMigrations(wanted, existing)
	if err != nil {
		return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
	}

	// execute missing
	for _, query := range missing {
		if _, err := pg.db.Exec(string(query)); err != nil {
			return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
		}

		// register
		if _, err := pg.db.Exec(`
INSERT INTO migration
(query) VALUES ($1)
`, query); err != nil {
			return fmt.Errorf("%w: %v", moviestore.ErrSqliteFailure, err)
		}
	}

	return nil
}

func (pg *Postgres) Exec(query string, args ...any) (sql.Result, error) {
	return pg.db.Exec(query, args...)
}

func (pg *Postgres) QueryRow(query string, args ...any) *sql.Row {
	return pg.db.QueryRow(query, args...)
}

func (pg *Postgres) Query(query string, args ...any) (*sql.Rows, error) {
	return pg.db.Query(query, args...)
}

func compareMigrations(wanted, existing []migration) ([]migration, error) {
	needed := []migration{}
	if len(wanted) < len(existing) {
		return []migration{}, moviestore.ErrNotEnoughSQLMigrations
	}

	for i, want := range wanted {
		switch {
		case i >= len(existing):
			needed = append(needed, want)
		case want == existing[i]:
			// do nothing
		case want != existing[i]:
			return []migration{}, fmt.Errorf("%w: %v", moviestore.ErrIncompatibleSQLMigration, want)
		}
	}

	return needed, nil
}
