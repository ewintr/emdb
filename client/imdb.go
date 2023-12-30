package client

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Review struct {
	Source string
	Review string
}

type IMDB struct {
}

func NewIMDB() *IMDB {
	return &IMDB{}
}

func (i *IMDB) GetReviews(imdbID string) (map[string]string, error) {
	url := fmt.Sprintf("https://www.imdb.com/title/%s/reviews", imdbID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	reviews := make(map[string]string)
	doc.Find(".lister-item-content").Each(func(i int, reviewNode *goquery.Selection) {

		var permaLink string
		reviewNode.Find("a").Each(func(i int, s *goquery.Selection) {
			if s.Text() == "Permalink" {
				link, exists := s.Attr("href")
				if exists {
					permaLink = link
				}
			}
		})

		if permaLink == "" {
			return
		}

		reviews[permaLink] = ScrubIMDBReview(reviewNode.Text())
	})

	return reviews, nil
}

func ScrubIMDBReview(review string) string {
	// remove footer
	for _, text := range []string{"Was this review helpful?", "Sign in to vote.", "Permalink"} {
		review = strings.ReplaceAll(review, text, "")
	}

	// remove superfluous whitespace
	reWS := regexp.MustCompile(`\n\s+`)
	review = reWS.ReplaceAllString(review, "\n")

	// remove superfluous newlines
	re := regexp.MustCompile(`\n{3,}`)
	review = re.ReplaceAllString(review, "\n\n")

	return review
}
