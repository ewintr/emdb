package worker

import "code.ewintr.nl/emdb/job"

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
		if err := w.jq.Add(review.ID, job.ActionFindTitles); err != nil {
			logger.Error("could not add job", "error", err)
			w.jq.MarkFailed(jobID)
			return
		}
	}

	logger.Info("refresh reviews", "count", len(reviews))
	w.jq.MarkDone(jobID)
}
