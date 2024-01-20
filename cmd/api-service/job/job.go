package job

import (
	"slices"
	"time"
)

type Status string
type Type string

const (
	interval = 20 * time.Second

	TypeSimple Type = "simple"
	TypeAI     Type = "ai"

	ActionRefreshIMDBReviews    = "refresh-imdb-reviews"
	ActionRefreshAllIMDBReviews = "refresh-all-imdb-reviews"
	ActionFindTitles            = "find-titles"
	ActionFindAllTitles         = "find-all-titles"
)

var (
	simpleActions = []string{
		ActionRefreshIMDBReviews,
		ActionRefreshAllIMDBReviews, // just creates a job for each movie
		ActionFindAllTitles,         // just creates a job for each review
	}
	aiActions = []string{
		ActionFindTitles,
	}

	validActions = append(simpleActions, aiActions...)
)

type Job struct {
	ID      int
	MovieID string
	Action  string
	Status  Status
	Created time.Time
	Updated time.Time
}

func Valid(action string) bool {
	if slices.Contains(validActions, action) {
		return true
	}

	return false
}
