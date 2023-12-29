package handler

import (
	"log/slog"

	"ewintr.nl/emdb/client"
	movie2 "ewintr.nl/emdb/cmd/api-service/moviestore"
	"github.com/google/uuid"
)

type Worker struct {
	jq         *movie2.JobQueue
	movieRepo  *movie2.MovieRepository
	reviewRepo *movie2.ReviewRepository
	imdb       *client.IMDB
	logger     *slog.Logger
}

func NewWorker(jq *movie2.JobQueue, movieRepo *movie2.MovieRepository, reviewRepo *movie2.ReviewRepository, imdb *client.IMDB, logger *slog.Logger) *Worker {
	return &Worker{
		jq:         jq,
		movieRepo:  movieRepo,
		reviewRepo: reviewRepo,
		imdb:       imdb,
		logger:     logger.With("service", "worker"),
	}
}

func (w *Worker) Run() {
	w.logger.Info("starting worker")
	for job := range w.jq.Next() {
		w.logger.Info("got a new job", "jobID", job.ID, "movieID", job.MovieID, "action", job.Action)
		switch job.Action {
		case movie2.ActionFetchIMDBReviews:
			w.fetchReviews(job)
		default:
			w.logger.Warn("unknown job action", "action", job.Action)
		}
	}
}

func (w *Worker) fetchReviews(job movie2.Job) {
	logger := w.logger.With("method", "fetchReviews", "jobID", job.ID, "movieID", job.MovieID)

	m, err := w.movieRepo.FindOne(job.MovieID)
	if err != nil {
		logger.Error("could not get movie", "error", err)
		return
	}

	reviews, err := w.imdb.GetReviews(m.IMDBID)
	if err != nil {
		logger.Error("could not get reviews", "error", err)
		return
	}

	for url, review := range reviews {
		if err := w.reviewRepo.Store(movie2.Review{
			ID:      uuid.New().String(),
			MovieID: m.ID,
			Source:  movie2.ReviewSourceIMDB,
			URL:     url,
			Review:  review,
		}); err != nil {
			logger.Error("could not store review", "error", err)
			return
		}
	}

	logger.Info("fetched reviews", "count", len(reviews))
}
