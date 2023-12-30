package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ewintr.nl/emdb/cmd/api-service/job"
)

type AdminAPI struct {
	jq     *job.JobQueue
	logger *slog.Logger
}

func NewAdminAPI(jq *job.JobQueue, logger *slog.Logger) *AdminAPI {
	return &AdminAPI{
		jq:     jq,
		logger: logger.With("api", "admin"),
	}
}

func (adminAPI *AdminAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := adminAPI.logger.With("method", "serveHTTP")

	subPath, _ := ShiftPath(r.URL.Path)
	switch {
	case r.Method == http.MethodPost && subPath == "":
		adminAPI.Add(w, r)
	default:
		Error(w, http.StatusNotFound, "unregistered path", nil, logger)
	}
}

func (adminAPI *AdminAPI) Add(w http.ResponseWriter, r *http.Request) {
	logger := adminAPI.logger.With("method", "add")

	var job job.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		Error(w, http.StatusBadRequest, "could not decode job", err, logger)
		return
	}

	if err := adminAPI.jq.Add(job.MovieID, job.Action); err != nil {
		Error(w, http.StatusInternalServerError, "could not add job", err, logger)
		return
	}

	if err := json.NewEncoder(w).Encode(job); err != nil {
		Error(w, http.StatusInternalServerError, "could not encode job", err, logger)
		return
	}
}
