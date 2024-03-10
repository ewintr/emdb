package job

import (
	"database/sql"
	"errors"
	"log/slog"

	"code.ewintr.nl/emdb/storage"
)

type JobQueue struct {
	db     *storage.Postgres
	logger *slog.Logger
}

func NewJobQueue(db *storage.Postgres, logger *slog.Logger) *JobQueue {
	jq := &JobQueue{
		db:     db,
		logger: logger.With("service", "jobqueue"),
	}

	return jq
}

func (jq *JobQueue) ResetAll() error {
	_, err := jq.db.Exec(`UPDATE job_queue SET status='todo'`)

	return err
}

func (jq *JobQueue) Add(movieID, action string) error {
	if !Valid(action) {
		return errors.New("invalid action")
	}

	_, err := jq.db.Exec(`
INSERT INTO job_queue (action_id, action, status) 
VALUES ($1, $2, 'todo');`, movieID, action)

	return err
}

func (jq *JobQueue) Next() (Job, error) {
	logger := jq.logger.With("method", "next")

	row := jq.db.QueryRow(`
SELECT id, action_id, action
FROM job_queue
WHERE status='todo'
ORDER BY id ASC
LIMIT 1;`)
	var job Job
	err := row.Scan(&job.ID, &job.ActionID, &job.Action)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error("could not fetch next job", "error", err)
		}
		return Job{}, err
	}

	logger.Info("found a job", "id", job.ID)
	if _, err := jq.db.Exec(`
UPDATE job_queue 
SET status='doing'
WHERE id=$1;`, job.ID); err != nil {
		logger.Error("could not set job to doing", "error")
		return Job{}, err
	}

	return job, nil
}

func (jq *JobQueue) MarkDone(id int) {
	logger := jq.logger.With("method", "markdone")
	if _, err := jq.db.Exec(`
DELETE FROM job_queue
WHERE id=$1;`, id); err != nil {
		logger.Error("could not mark job done", "error", err)
	}
	return
}

func (jq *JobQueue) MarkFailed(id int) {
	logger := jq.logger.With("method", "markfailed")
	if _, err := jq.db.Exec(`
UPDATE job_queue
SET status='failed'
WHERE id=$1;`, id); err != nil {
		logger.Error("could not mark job failed", "error", err)
	}
	return
}

func (jq *JobQueue) List() ([]Job, error) {
	rows, err := jq.db.Query(`
SELECT id, action_id, action, status, created_at, updated_at
FROM job_queue
ORDER BY id DESC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
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
WHERE id=$1;`, id); err != nil {
		return err
	}
	return nil
}

func (jq *JobQueue) DeleteAll() error {
	if _, err := jq.db.Exec(`DELETE FROM job_queue;`); err != nil {
		return err
	}
	return nil
}
