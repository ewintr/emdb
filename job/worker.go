package job

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"code.ewintr.nl/emdb/client"
	"code.ewintr.nl/emdb/storage"
)

type Worker struct {
	jq         *JobQueue
	movieRepo  *storage.MovieRepository
	reviewRepo *storage.ReviewRepository
	imdb       *client.IMDB
	logger     *slog.Logger
}

func NewWorker(jq *JobQueue, movieRepo *storage.MovieRepository, reviewRepo *storage.ReviewRepository, imdb *client.IMDB, logger *slog.Logger) *Worker {
	return &Worker{
		jq:         jq,
		movieRepo:  movieRepo,
		reviewRepo: reviewRepo,
		imdb:       imdb,
		logger:     logger.With("service", "worker"),
	}
}

func (w *Worker) Run() {
	logger := w.logger.With("method", "run")
	logger.Info("starting worker")
	for {
		time.Sleep(interval)
		j, err := w.jq.Next()
		switch {
		case errors.Is(err, sql.ErrNoRows):
			logger.Info("no simple jobs found")
			continue
		case err != nil:
			logger.Error("could not get next job", "error", err)
			continue
		}

		logger.Info("got a new job", "jobID", j.ID, "movieID", j.ActionID, "action", j.Action)
		switch j.Action {
		case ActionRefreshIMDBReviews:
			w.RefreshReviews(j.ID, j.ActionID)
		case ActionRefreshAllIMDBReviews:
			w.RefreshAllReviews(j.ID)
		case ActionFindAllTitles:
			w.FindAllTitles(j.ID)
		default:
			logger.Error("unknown job action", "action", j.Action)
		}
	}
}

func (w *Worker) RefreshAllReviews(jobID int) {
	logger := w.logger.With("method", "fetchReviews", "jobID", jobID)

	movies, err := w.movieRepo.FindAll()
	if err != nil {
		logger.Error("could not get movies", "error", err)
		return
	}

	for _, m := range movies {
		time.Sleep(1 * time.Second)
		if err := w.jq.Add(m.ID, ActionRefreshIMDBReviews); err != nil {
			logger.Error("could not add job", "error", err)
			return
		}
	}

	logger.Info("refresh all reviews", "count", len(movies))
	w.jq.MarkDone(jobID)
}

func (w *Worker) FindAllTitles(jobID int) {
	logger := w.logger.With("method", "findTitles", "jobID", jobID)

	reviews, err := w.reviewRepo.FindAll()
	if err != nil {
		logger.Error("could not get reviews", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	for _, r := range reviews {
		time.Sleep(1 * time.Second)
		if err := w.jq.Add(r.ID, ActionFindTitles); err != nil {
			logger.Error("could not add job", "error", err)
			w.jq.MarkFailed(jobID)
			return
		}
	}

	logger.Info("find all titles", "count", len(reviews))
	w.jq.MarkDone(jobID)
}

func (w *Worker) RefreshReviews(jobID int, movieID string) {
	logger := w.logger.With("method", "fetchReviews", "jobID", jobID, "movieID", movieID)

	m, err := w.movieRepo.FindOne(movieID)
	if err != nil {
		logger.Error("could not get movie", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	if err := w.reviewRepo.DeleteByMovieID(m.ID); err != nil {
		logger.Error("could not delete reviews", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	reviews, err := w.imdb.GetReviews(m)
	if err != nil {
		logger.Error("could not get reviews", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	for _, review := range reviews {
		if err := w.reviewRepo.Store(review); err != nil {
			logger.Error("could not store review", "error", err)
			w.jq.MarkFailed(jobID)
			return
		}
		if err := w.jq.Add(review.ID, ActionFindTitles); err != nil {
			logger.Error("could not add job", "error", err)
			w.jq.MarkFailed(jobID)
			return
		}
	}

	logger.Info("refresh reviews", "count", len(reviews))
	w.jq.MarkDone(jobID)
}
