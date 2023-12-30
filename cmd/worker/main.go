package main

import (
	"context"
	"fmt"
	"os"

	"ewintr.nl/emdb/client"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	//movieID := "c313ffa9-1cec-4d37-be6e-a4e18c247a0f" // night train
	movieID := "07fb2a59-24ec-442e-aa9e-eb7e4fb002db" // battle royale

	emdb := client.NewEMDB(os.Getenv("EMDB_BASE_URL"), os.Getenv("EMDB_API_KEY"))
	reviews, err := emdb.GetReviews(movieID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	llm, err := ollama.New(ollama.WithModel("dolphin-mixtral"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx := context.Background()

	for _, review := range reviews {
		fmt.Printf("\nReview: %s\nAnswer:\n", review.Review)

		prompt := fmt.Sprintf(`Human: The following is a comment about the movie "Battle Royale": 
		---%s
		---What other movies are referenced in this comment? If you're not sure, just say so. Don't make any other comment. Just give a list of movies if there are any.
		Assistant:`, review.Review)
		_, err := llm.Call(ctx, prompt,
			llms.WithTemperature(0.8),
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Print(string(chunk))
				return nil
			}),
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
