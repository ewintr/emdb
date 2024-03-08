package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"code.ewintr.nl/emdb/client"
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

func main() {
	emdb := client.NewEMDB(os.Getenv("EMDB_BASE_URL"), os.Getenv("EMDB_API_KEY"))

	go Work(emdb)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

func Work(emdb *client.EMDB) {
	for {
		j, err := emdb.GetNextAIJob()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		review, err := emdb.GetReview(j.ActionID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		movie, err := emdb.GetMovie(review.MovieID)
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
			os.Exit(1)
		}

		review.Titles = resp

		if err := emdb.UpdateReview(review); err != nil {
			fmt.Printf("could not update review: %s\n", err)
			os.Exit(1)
		}
	}
}
