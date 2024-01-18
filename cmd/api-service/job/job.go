package job

import (
	"time"
)

type JobStatus string

type Action string

const (
	interval = 10 * time.Second

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
