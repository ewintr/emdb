package job

import (
	"slices"
	"time"
)

type JobStatus string
type JobType string

const (
	TypeSimple JobType = "simple"
	TypeAI     JobType = "ai"

	ActionRefreshIMDBReviews    = "refresh-imdb-reviews"
	ActionRefreshAllIMDBReviews = "refresh-all-imdb-reviews"
	ActionFindTitles            = "find-titles"
	ActionFindAllTitles         = "find-all-titles"
)

var (
	SimpleActions = []string{
		ActionRefreshIMDBReviews,
		ActionRefreshAllIMDBReviews, // just creates a job for each movie
		ActionFindAllTitles,         // just creates a job for each review
	}
	AIActions = []string{
		ActionFindTitles,
	}

	ValidActions = append(SimpleActions, AIActions...)
)

type Job struct {
	ID       int
	ActionID string
	Action   string
	Status   JobStatus
	Created  time.Time
	Updated  time.Time
}

func Valid(action string) bool {
	if slices.Contains(ValidActions, action) {
		return true
	}

	return false
}
