package main

import (
	"fmt"
	"sync"
)

// ApplicantRecord is app-owned business data.
type ApplicantRecord struct {
	ApplicantID string
	RunID       string
	Profile     Profile
	Documents   []Document
}

type Profile struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

type Document struct {
	Type string `json:"documentType"`
	Ref  string `json:"documentRef"`
}

type ApplicantStore struct {
	mu          sync.Mutex
	byRunID     map[string]*ApplicantRecord
	byApplicant map[string]*ApplicantRecord
}

func NewApplicantStore() *ApplicantStore {
	return &ApplicantStore{
		byRunID:     make(map[string]*ApplicantRecord),
		byApplicant: make(map[string]*ApplicantRecord),
	}
}

func (s *ApplicantStore) Create(applicantID, runID string) *ApplicantRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec := &ApplicantRecord{ApplicantID: applicantID, RunID: runID}
	s.byRunID[runID] = rec
	s.byApplicant[applicantID] = rec
	return rec
}

func (s *ApplicantStore) GetByRunID(runID string) (*ApplicantRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.byRunID[runID]
	if !ok {
		return nil, fmt.Errorf("%w: run %q", ErrApplicantNotFound, runID)
	}
	return cloneApplicant(rec), nil
}

func (s *ApplicantStore) SaveProfile(runID string, p Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.byRunID[runID]
	if !ok {
		return fmt.Errorf("%w: run %q", ErrApplicantNotFound, runID)
	}
	rec.Profile = p
	return nil
}

func (s *ApplicantStore) AddDocument(runID string, d Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.byRunID[runID]
	if !ok {
		return fmt.Errorf("%w: run %q", ErrApplicantNotFound, runID)
	}
	rec.Documents = append(rec.Documents, d)
	return nil
}

func cloneApplicant(rec *ApplicantRecord) *ApplicantRecord {
	cp := *rec
	cp.Documents = append([]Document(nil), rec.Documents...)
	return &cp
}
