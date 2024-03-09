package job

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"code.ewintr.nl/emdb/cmd/api-service/moviestore"
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

	go jq.Run()

	return jq
}

func (jq *JobQueue) Run() {
	logger := jq.logger.With("method", "run")
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-ticker.C:
			logger.Info("resetting stuck jobs")
			if _, err := jq.db.Exec(`
UPDATE job_queue 
SET status = 'todo'
WHERE status = 'doing' 
	AND EXTRACT(EPOCH FROM now() - updated_at) > 2*24*60*60;`); err != nil {
				logger.Error("could not clean up job queue", "error", err)
			}
		}
	}
}

func (jq *JobQueue) Add(movieID, action string) error {
	if !moviestore.Valid(action) {
		return errors.New("invalid action")
	}

	_, err := jq.db.Exec(`
INSERT INTO job_queue (action_id, action, status) 
VALUES ($1, $2, 'todo');`, movieID, action)

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
	AND action = ANY($1)
ORDER BY id ASC
LIMIT 1;`, actionsStr)
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
WHERE id=$1;`, job.ID); err != nil {
		logger.Error("could not set job to doing", "error")
		return moviestore.Job{}, err
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

func (jq *JobQueue) List() ([]moviestore.Job, error) {
	rows, err := jq.db.Query(`
SELECT id, action_id, action, status, created_at, updated_at
FROM job_queue
ORDER BY id DESC;`)
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
