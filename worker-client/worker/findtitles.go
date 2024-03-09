package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

const (
	mentionsTemplate = `The following text is a user comment about the movie {{.title}}. In it, the user may have referenced other movie titles. List them if you see any.

----
{{.review}}
---- 

If you found any movie titles other than {{.title}}, list them below in a JSON array. If there are other titles, like TV shows, books or games, ignore them. The format is as follows:

["movie title 1", "movie title 2"]

Just answer with the JSON and nothing else. If you don't see any other movie titles, just answer with an empty array.`
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
	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		logger.Error("could not create llm", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	ctx := context.Background()

	prompt := prompts.NewPromptTemplate(
		mentionsTemplate,
		[]string{"title", "review"},
	)
	llmChain := chains.NewLLMChain(llm, prompt)

	movieTitle := movie.Title
	if movie.EnglishTitle != "" && movie.EnglishTitle != movie.Title {
		movieTitle = fmt.Sprintf("%s (English title: %s)", movieTitle, movie.EnglishTitle)
	}
	fmt.Printf("Processing review for movie: %s\n", movieTitle)
	fmt.Printf("Review: %s\n", review.Review)

	outputValues, err := chains.Call(ctx, llmChain, map[string]any{
		"title":  movieTitle,
		"review": review.Review,
	})
	if err != nil {
		logger.Error("could not call chain", "error", err)
		w.jq.MarkFailed(jobID)
		return
	}

	out, ok := outputValues[llmChain.OutputKey].(string)
	if !ok {
		logger.Error("chain output is not valid")
		w.jq.MarkFailed(jobID)
		return
	}
	//fmt.Println(out)
	resp := struct {
		Movies  []string `json:"movies"`
		TVShows []string `json:"tvShows"`
		Games   []string `json:"games"`
		Books   []string `json:"books"`
	}{}

	if err := json.Unmarshal([]byte(out), &resp); err != nil {
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
