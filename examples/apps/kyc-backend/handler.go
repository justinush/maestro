package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/justinush/maestro/pkg/run"
)

type Server struct {
	svc *KYCService
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /kyc/start", s.handleStart)
	mux.HandleFunc("GET /kyc/{runID}", s.handleGet)
	mux.HandleFunc("GET /kyc/{runID}/events", s.handleEvents)
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

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runID")
	resp, err := s.svc.GetEvents(r.Context(), runID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("runID")
	var body Profile
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid json: trailing data")
	}
	return nil
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
	case errors.Is(err, ErrApplicantNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrWrongStep):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
