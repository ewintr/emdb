package worker

import (
	"encoding/json"
	"fmt"
)

const (
	mentionsTemplate = `The following text is a user comment about the movie %s. In it, the user may have referenced other movie titles. List them if you see any.

----
%s
---- 

If you found any movie titles other than %s, list them below in a JSON array. If there are other titles, like TV shows, books or games, ignore them. The format is as follows:

["movie title 1", "movie title 2"]

Just answer with the JSON and nothing else. If you don't see any other movie titles, just answer with an empty JSON array.`
)

func (w *Worker) FindTitles(jobID int, reviewID string) {
	logger := w.logger.With("method", "findTitles", "jobID", jobID)

	review, err := w.reviewRepo.FindOne(reviewID)
	if err != nil {
		logger.Error("could not get review", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	movie, err := w.movieRepo.FindOne(review.MovieID)
	if err != nil {
		logger.Error("could not get movie", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	movieTitle := movie.Title
	if movie.EnglishTitle != "" && movie.EnglishTitle != movie.Title {
		movieTitle = fmt.Sprintf("%s (English title: %s)", movieTitle, movie.EnglishTitle)
	}

	prompt := fmt.Sprintf(mentionsTemplate, movieTitle, review.Review, movieTitle)
	titles, err := w.ollama.Generate("mistral", prompt)
	if err != nil {
		logger.Error("could not find titles: %w", err)
	}
	logger.Info("checked review", "found", titles)
	var resp []string
	if err := json.Unmarshal([]byte(titles), &resp); err != nil {
		logger.Error("could not unmarshal llm response", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	review.Titles = resp

	if err := w.reviewRepo.Store(review); err != nil {
		logger.Error("could not update review", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}
}
