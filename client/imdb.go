package client

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"ewintr.nl/emdb/cmd/api-service/moviestore"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

type IMDB struct {
}

func NewIMDB() *IMDB {
	return &IMDB{}
}

func (i *IMDB) GetReviews(m moviestore.Movie) ([]moviestore.Review, error) {
	url := fmt.Sprintf("https://www.imdb.com/title/%s/reviews", m.IMDBID)
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

	reviews := make([]moviestore.Review, 0)
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

		rat, rev := ScrubIMDBReview(reviewNode.Text())
		reviews = append(reviews, moviestore.Review{
			ID:          uuid.New().String(),
			MovieID:     m.ID,
			Source:      moviestore.ReviewSourceIMDB,
			URL:         fmt.Sprintf("https://www.imdb.com%s", permaLink),
			Review:      rev,
			MovieRating: rat,
		})
	})

	return reviews, nil
}

func ScrubIMDBReview(review string) (int, string) {
	// remove footer
	for _, text := range []string{"Was this review helpful?", "Sign in to vote.", "Permalink"} {
		review = strings.ReplaceAll(review, text, "")
	}

	// remove superfluous whitespace
	reWS := regexp.MustCompile(`\n\s+`)
	review = reWS.ReplaceAllString(review, "\n")

	// remove superfluous newlines
	reRev := regexp.MustCompile(`\n{3,}`)
	review = reRev.ReplaceAllString(review, "\n\n")

	reRat := regexp.MustCompile(`(\d+)/10\n`)
	reMatch := reRat.FindStringSubmatch(review)
	var rating int
	if len(reMatch) > 0 {
		rating, _ = strconv.Atoi(reMatch[1])
		review = strings.ReplaceAll(review, reMatch[0], "")
	}

	return rating, review
}
