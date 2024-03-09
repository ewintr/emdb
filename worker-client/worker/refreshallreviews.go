package worker

import (
	"time"

	"code.ewintr.nl/emdb/job"
)

func (w *Worker) RefreshAllReviews(jobID int) {
	logger := w.logger.With("method", "fetchReviews", "jobID", jobID)

	movies, err := w.movieRepo.FindAll()
	if err != nil {
		logger.Error("could not get movies", "error", err)
		return
	}

	for _, m := range movies {
		time.Sleep(1 * time.Second)
		if err := w.jq.Add(m.ID, job.ActionRefreshIMDBReviews); err != nil {
			logger.Error("could not add job", "error", err)
			return
		}
	}

	logger.Info("refresh all reviews", "count", len(movies))
	w.jq.MarkDone(jobID)
}
