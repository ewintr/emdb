package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
)

type APIIndex map[string]http.Handler

type Server struct {
	apiKey string
	apis   map[string]http.Handler
	logger *slog.Logger
}

func NewServer(apiKey string, apis map[string]http.Handler, logger *slog.Logger) *Server {
	return &Server{
		apiKey: apiKey,
		apis:   apis,
		logger: logger,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := httptest.NewRecorder() // records the response to be able to mix writing headers and content

	// cors
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		returnResponse(w, rec, r, s.logger)
		return
	}

	logger := s.logger.With("path", r.URL.Path)
	w.Header().Add("Content-Type", "application/json")

	// authenticate
	if key := r.Header.Get("Authorization"); s.apiKey != "localOnly" && key != s.apiKey {
		Error(rec, http.StatusUnauthorized, "unauthorized", fmt.Errorf("invalid api key"), logger)
		logger.Info("unauthorized", "key", key)
		returnResponse(w, rec, r, logger)
		return
	}

	// route to internal
	head, tail := ShiftPath(r.URL.Path)
	if len(head) == 0 {
		Index(rec)
		returnResponse(w, rec, r, logger)
		return
	}
	api, ok := s.apis[head]
	if !ok {
		Error(rec, http.StatusNotFound, "Not found", fmt.Errorf("%s is not a valid path", r.URL.Path), logger)
		returnResponse(w, rec, r, logger)
		return
	}

	r.URL.Path = tail
	api.ServeHTTP(rec, r)
	returnResponse(w, rec, r, logger)
}

func returnResponse(w http.ResponseWriter, rec *httptest.ResponseRecorder, r *http.Request, logger *slog.Logger) {
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
	logger.Info("request served", "method", r.Method, "status", rec.Code)
}

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
// See https://blog.merovius.de/posts/2017-06-18-how-not-to-use-an-http-router/
func ShiftPath(p string) (string, string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
