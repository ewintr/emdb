package worker

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/job"
	"code.ewintr.nl/emdb/storage"
)

const (
	interval = 5 * time.Second
)

type Worker struct {
	jq         *job.JobQueue
	movieRepo  *storage.MovieRepository
	reviewRepo *storage.ReviewRepository
	imdb       *client.IMDB
	ollama     *client.Ollama
	logger     *slog.Logger
}

func NewWorker(jq *job.JobQueue, movieRepo *storage.MovieRepository, reviewRepo *storage.ReviewRepository, imdb *client.IMDB, ollama *client.Ollama, logger *slog.Logger) *Worker {
	return &Worker{
		jq:         jq,
		movieRepo:  movieRepo,
		reviewRepo: reviewRepo,
		imdb:       imdb,
		ollama:     ollama,
		logger:     logger.With("service", "worker"),
	}
}

func (w *Worker) Run() {
	logger := w.logger.With("method", "run")
	logger.Info("starting worker")

	logger.Info("setting al existing jobs to todo")
	if err := w.jq.ResetAll(); err != nil {
		logger.Error("could not set all jobs to todo", "error", err)
		return
	}

	for {
		time.Sleep(interval)

		j, err := w.jq.Next()
		switch {
		case errors.Is(err, sql.ErrNoRows):
			//logger.Info("no jobs found")
			continue
		case err != nil:
			logger.Error("could not get next job", "error", err)
			continue
		}

		logger.Info("got a new job", "jobID", j.ID, "movieID", j.ActionID, "action", j.Action)
		switch j.Action {
		case job.ActionRefreshIMDBReviews:
			w.RefreshReviews(j.ID, j.ActionID)
		case job.ActionRefreshAllIMDBReviews:
			w.RefreshAllReviews(j.ID)
		case job.ActionFindTitles:
			w.FindTitles(j.ID, j.ActionID)
		case job.ActionFindAllTitles:
			w.FindAllTitles(j.ID)
		default:
			logger.Error("unknown job action", "action", j.Action)
		}
	}
}
