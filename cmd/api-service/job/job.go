package job

import (
	"time"
)

type JobStatus string

const (
	JobStatusToDo  JobStatus = "todo"
	JobStatusDoing JobStatus = "doing"
	JobStatusDone  JobStatus = "done"
)

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
