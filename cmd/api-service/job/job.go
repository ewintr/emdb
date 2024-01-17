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

type Job struct {
	ID      int
	MovieID string
	Action  Action
	Status  JobStatus
}
