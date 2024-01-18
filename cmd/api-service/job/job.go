package job

import (
	"slices"
	"time"
)

type JobStatus string

type Action string

const (
	interval = 20 * time.Second

	ActionRefreshIMDBReviews    Action = "refresh-imdb-reviews"
	ActionRefreshAllIMDBReviews Action = "refresh-all-imdb-reviews"
)

var (
	validActions = []Action{
		ActionRefreshIMDBReviews,
		ActionRefreshAllIMDBReviews,
	}
)

type Job struct {
	ID      int
	MovieID string
	Action  Action
	Status  JobStatus
	Created time.Time
	Updated time.Time
}

func Valid(action Action) bool {
	if slices.Contains(validActions, action) {
		return true
	}

	return false
}
