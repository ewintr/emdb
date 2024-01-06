package job

import (
	"log/slog"

	"ewintr.nl/emdb/client"
	"ewintr.nl/emdb/cmd/api-service/moviestore"
	"github.com/google/uuid"
)

type Worker struct {
	jq         *JobQueue
	movieRepo  *moviestore.MovieRepository
	reviewRepo *moviestore.ReviewRepository
	imdb       *client.IMDB
	logger     *slog.Logger
}

func NewWorker(jq *JobQueue, movieRepo *moviestore.MovieRepository, reviewRepo *moviestore.ReviewRepository, imdb *client.IMDB, logger *slog.Logger) *Worker {
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
	for j := range w.jq.Next() {
		w.logger.Info("got a new job", "jobID", j.ID, "movieID", j.MovieID, "action", j.Action)
		switch j.Action {
		case ActionRefreshIMDBReviews:
			w.RefreshReviews(j.ID, j.MovieID)
		case ActionRefreshAllIMDBReviews:
			w.RefreshAllReviews(j.ID)
		default:
			w.logger.Warn("unknown job action", "action", j.Action)
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
		if err := w.jq.Add(m.ID, ActionRefreshIMDBReviews); err != nil {
			logger.Error("could not add job", "error", err)
			return
		}
	}
}

func (w *Worker) RefreshReviews(jobID int, movieID string) {
	logger := w.logger.With("method", "fetchReviews", "jobID", jobID, "movieID", movieID)

	m, err := w.movieRepo.FindOne(movieID)
	if err != nil {
		logger.Error("could not get movie", "error", err)
		return
	}

	if err := w.reviewRepo.DeleteByMovieID(m.ID); err != nil {
		logger.Error("could not delete reviews", "error", err)
		return
	}

	reviews, err := w.imdb.GetReviews(m.IMDBID)
	if err != nil {
		logger.Error("could not get reviews", "error", err)
		return
	}

	for url, review := range reviews {
		if err := w.reviewRepo.Store(moviestore.Review{
			ID:      uuid.New().String(),
			MovieID: m.ID,
			Source:  moviestore.ReviewSourceIMDB,
			URL:     url,
			Review:  review,
		}); err != nil {
			logger.Error("could not store review", "error", err)
			return
		}
	}

	logger.Info("refresh reviews", "count", len(reviews))
}
