package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/justinush/maestro/pkg/run"
)

type Server struct {
	svc *KYCService
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /kyc/start", s.handleStart)
	mux.HandleFunc("GET /kyc/{runID}", s.handleGet)
	mux.HandleFunc("POST /kyc/{runID}/profile", s.handleProfile)
	mux.HandleFunc("POST /kyc/{runID}/document", s.handleDocument)
	mux.HandleFunc("POST /kyc/{runID}/review", s.handleReview)
	return mux
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	resp, err := s.svc.Start(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runID")
	resp, err := s.svc.Get(r.Context(), runID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runID")
	var body Profile
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	resp, err := s.svc.SubmitProfile(r.Context(), runID, body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDocument(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runID")
	var body Document
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	resp, err := s.svc.SubmitDocument(r.Context(), runID, body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleReview(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runID")
	var body struct {
		Approved bool `json:"approved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if !body.Approved {
		http.Error(w, "demo only supports approved=true", http.StatusBadRequest)
		return
	}
	resp, err := s.svc.SubmitReview(r.Context(), runID, body.Approved)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(data); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, run.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		msg := err.Error()
		if strings.Contains(msg, "wrong step:") {
			http.Error(w, msg, http.StatusConflict)
			return
		}
		if strings.Contains(msg, "applicant:") {
			http.Error(w, msg, http.StatusNotFound)
			return
		}
		http.Error(w, msg, http.StatusInternalServerError)
	}
}
