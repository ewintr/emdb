package job

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"ewintr.nl/emdb/cmd/api-service/moviestore"
)

type JobQueue struct {
	db     *moviestore.SQLite
	logger *slog.Logger
}

func NewJobQueue(db *moviestore.SQLite, logger *slog.Logger) *JobQueue {
	return &JobQueue{
		db:     db,
		logger: logger.With("service", "jobqueue"),
	}
}

func (jq *JobQueue) Add(movieID, action string) error {
	if !moviestore.Valid(action) {
		return errors.New("invalid action")
	}

	_, err := jq.db.Exec(`INSERT INTO job_queue (action_id, action, status) 
	VALUES (?, ?, 'todo')`, movieID, action)

	return err
}

func (jq *JobQueue) Next(t moviestore.JobType) (moviestore.Job, error) {
	logger := jq.logger.With("method", "next")

	actions := moviestore.SimpleActions
	if t == moviestore.TypeAI {
		actions = moviestore.AIActions
	}
	actionsStr := fmt.Sprintf("('%s')", strings.Join(actions, "', '"))
	query := fmt.Sprintf(`
SELECT id, action_id, action
FROM job_queue
WHERE status='todo'
	AND action IN %s
ORDER BY id ASC
LIMIT 1`, actionsStr)
	row := jq.db.QueryRow(query)
	var job moviestore.Job
	err := row.Scan(&job.ID, &job.ActionID, &job.Action)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error("could not fetch next job", "error", err)
		}
		return moviestore.Job{}, err
	}

	logger.Info("found a job", "id", job.ID)
	if _, err := jq.db.Exec(`
UPDATE job_queue 
SET status='doing'
WHERE id=?`, job.ID); err != nil {
		logger.Error("could not set job to doing", "error")
		return moviestore.Job{}, err
	}

	return job, nil
}

func (jq *JobQueue) MarkDone(id int) {
	logger := jq.logger.With("method", "markdone")
	if _, err := jq.db.Exec(`
DELETE FROM job_queue
WHERE id=?`, id); err != nil {
		logger.Error("could not mark job done", "error", err)
	}
	return
}

func (jq *JobQueue) MarkFailed(id int) {
	logger := jq.logger.With("method", "markfailed")
	if _, err := jq.db.Exec(`
UPDATE job_queue
SET status='failed'
WHERE id=?`, id); err != nil {
		logger.Error("could not mark job failed", "error", err)
	}
	return
}

func (jq *JobQueue) List() ([]moviestore.Job, error) {
	rows, err := jq.db.Query(`
SELECT id, action_id, action, status, created_at, updated_at
FROM job_queue
ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []moviestore.Job
	for rows.Next() {
		var j moviestore.Job
		if err := rows.Scan(&j.ID, &j.ActionID, &j.Action, &j.Status, &j.Created, &j.Updated); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (jq *JobQueue) Delete(id string) error {
	if _, err := jq.db.Exec(`
DELETE FROM job_queue
WHERE id=?`, id); err != nil {
		return err
	}
	return nil
}

func (jq *JobQueue) DeleteAll() error {
	if _, err := jq.db.Exec(`
DELETE FROM job_queue`); err != nil {
		return err
	}
	return nil
}
