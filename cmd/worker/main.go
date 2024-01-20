package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"ewintr.nl/emdb/client"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

const (
	mentionsTemplate = `The following text is a user comment about the movie {{.title}}. In it, the user may have referenced other movie titles. List them if you see any.

----
{{.review}}
---- 

If you found any movie titles other than {{.title}}, list them below in a JSON array. If there are other titles, like TV shows, books or games, separate them according to the following schema:

{
"movies": ["movie title 1", "movie title 2"],
"tvShows": ["tv series 1"],
"games": ["game 1", "game 2"],
"books": ["book"]
}

Just answer with the JSON and nothing else. When in doubt about whether a title is a movie or another category, don't put it in movies, but in the other category.  `
)

func main() {
	//movieID := "c313ffa9-1cec-4d37-be6e-a4e18c247a0f" // night train
	//movieID := "07fb2a59-24ec-442e-aa9e-eb7e4fb002db" // battle royale
	movieID := "2fce2f8f-a048-4e39-8ffe-82df09a29d32" // shadows in paradise

	emdb := client.NewEMDB(os.Getenv("EMDB_BASE_URL"), os.Getenv("EMDB_API_KEY"))
	movie, err := emdb.GetMovie(movieID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	reviews, err := emdb.GetReviews(movieID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := context.Background()

	prompt := prompts.NewPromptTemplate(
		mentionsTemplate,
		[]string{"title", "review"},
	)
	llmChain := chains.NewLLMChain(llm, prompt)

	fmt.Printf("Processing review for movie: %s\n", movie.Title)
	for _, review := range reviews {
		fmt.Printf("Review: %s\n", review.Review)
		outputValues, err := chains.Call(ctx, llmChain, map[string]any{
			"title":  movie.Title,
			"review": review.Review,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		out, ok := outputValues[llmChain.OutputKey].(string)
		if !ok {
			fmt.Println("invalid chain return")
		}
		fmt.Println(out)
		resp := struct {
			Movies  []string `json:"movies"`
			TVShows []string `json:"tvShows"`
			Games   []string `json:"games"`
			Books   []string `json:"books"`
		}{}

		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			fmt.Printf("could not unmarshal llm response, skipping this one: %s", err)
			continue
		}

		fmt.Printf("Movies: %v\n", resp.Movies)

		review.Titles = resp

		if err := emdb.UpdateReview(review); err != nil {
			fmt.Printf("could not update review: %s\n", err)
			continue
		}

	}
}
