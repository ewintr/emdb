package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ewintr.nl/emdb/cmd/api-service/job"
)

type JobAPI struct {
	jq     *job.JobQueue
	logger *slog.Logger
}

func NewJobAPI(jq *job.JobQueue, logger *slog.Logger) *JobAPI {
	return &JobAPI{
		jq:     jq,
		logger: logger.With("api", "admin"),
	}
}

func (jobAPI *JobAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := jobAPI.logger.With("method", "serveHTTP")

	subPath, _ := ShiftPath(r.URL.Path)
	switch {
	case r.Method == http.MethodPost && subPath == "":
		jobAPI.Add(w, r)
	case r.Method == http.MethodGet && subPath == "":
		jobAPI.List(w, r)
	case r.Method == http.MethodDelete && subPath != "":
		jobAPI.Delete(w, r, subPath)
	case r.Method == http.MethodDelete && subPath == "":
		jobAPI.DeleteAll(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", nil, logger)
	}
}

func (jobAPI *JobAPI) Add(w http.ResponseWriter, r *http.Request) {
	logger := jobAPI.logger.With("method", "add")

	var j job.Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		Error(w, http.StatusBadRequest, "could not decode job", err, logger)
		return
	}

	if err := jobAPI.jq.Add(j.MovieID, j.Action); err != nil {
		Error(w, http.StatusInternalServerError, "could not add job", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(j); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode job", err, logger)
		return
	}
}

func (jobAPI *JobAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := jobAPI.logger.With("method", "list")

	jobs, err := jobAPI.jq.List()
	if err != nil {
		Error(w, http.StatusInternalServerError, "could not list jobs", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode jobs", err, logger)
		return
	}
}

func (jobAPI *JobAPI) Delete(w http.ResponseWriter, r *http.Request, id string) {
	logger := jobAPI.logger.With("method", "delete")

	if err := jobAPI.jq.Delete(id); err != nil {
		Error(w, http.StatusInternalServerError, "could not delete job", err, logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (jobAPI *JobAPI) DeleteAll(w http.ResponseWriter, r *http.Request) {
	logger := jobAPI.logger.With("method", "deleteall")

	if err := jobAPI.jq.DeleteAll(); err != nil {
		Error(w, http.StatusInternalServerError, "could not delete all jobs", err, logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
