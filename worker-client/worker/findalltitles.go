package worker

import (
	"go-mod.ewintr.nl/emdb/job"
)

func (w *Worker) FindAllTitles(jobID int) {
	logger := w.logger.With("method", "findAllTitles", "jobID", jobID)

	reviews, err := w.reviewRepo.FindAll()
	if err != nil {
		logger.Error("could not get reviews", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	for _, r := range reviews {
		if err := w.jq.Add(r.ID, job.ActionFindTitles); err != nil {
			logger.Error("could not add job", "error", err)
			w.jq.MarkFailed(jobID)
			return
		}
	}

	logger.Info("find all titles", "count", len(reviews))
	w.jq.MarkDone(jobID)
}
