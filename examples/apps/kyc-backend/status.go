package main

import "github.com/justinush/maestro/pkg/engine"

// UI-facing status derived from Maestro step + run outcome.

type StatusResponse struct {
	RunID       string     `json:"runId"`
	ApplicantID string     `json:"applicantId"`
	Status      string     `json:"status"`
	Step        string     `json:"step"`
	Terminal    bool       `json:"terminal"`
	Profile     *Profile   `json:"profile,omitempty"`
	Documents   []Document `json:"documents,omitempty"`
}

func buildStatus(applicant *ApplicantRecord, in *engine.Instance, completed bool) StatusResponse {
	step := in.CurrentStepID()
	resp := StatusResponse{
		RunID:       applicant.RunID,
		ApplicantID: applicant.ApplicantID,
		Step:        step,
		Terminal:    completed || in.IsTerminal(),
		Status:      mapStepToStatus(step, completed || in.IsTerminal()),
	}
	if applicant.Profile.FullName != "" || applicant.Profile.Email != "" {
		p := applicant.Profile
		resp.Profile = &p
	}
	if len(applicant.Documents) > 0 {
		resp.Documents = applicant.Documents
	}
	return resp
}

func mapStepToStatus(stepID string, terminal bool) string {
	if terminal && stepID == "approved" {
		return "approved"
	}
	switch stepID {
	case "collect-profile":
		return "awaiting_profile"
	case "document-upload":
		return "awaiting_document"
	case "run-liveness-check":
		return "processing_liveness"
	case "manual-review":
		return "awaiting_review"
	default:
		return "in_progress"
	}
}
