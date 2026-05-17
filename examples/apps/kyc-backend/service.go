package main

import (
	"context"
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
)

type KYCService struct {
	rt         *maestro.Runtime
	def        *definition.WorkflowDefinition
	runs       run.Store
	applicants *ApplicantStore
}

func NewKYCService(rt *maestro.Runtime, runs run.Store, applicants *ApplicantStore) *KYCService {
	return &KYCService{
		rt:         rt,
		def:        rt.Definition(),
		runs:       runs,
		applicants: applicants,
	}
}

func (s *KYCService) Start(ctx context.Context) (StatusResponse, error) {
	applicantID := newID("app")
	runID := newID("run")

	s.applicants.Create(applicantID, runID)

	in, err := s.rt.NewInstance(maestro.InstanceOptions{
		RunID: runID,
		InitialVariables: map[string]any{
			"applicantId": applicantID,
		},
	})
	if err != nil {
		return StatusResponse{}, err
	}

	if err := s.driveUntilBlocked(in); err != nil {
		return StatusResponse{}, err
	}
	if err := persistNewRun(ctx, s.runs, in, s.def); err != nil {
		return StatusResponse{}, err
	}

	app, err := s.applicants.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	return buildStatus(app, in, false), nil
}

func (s *KYCService) Get(ctx context.Context, runID string) (StatusResponse, error) {
	app, err := s.applicants.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	rec, err := s.runs.Get(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	in, err := s.rt.RestoreInstance(rec, maestro.InstanceOptions{})
	if err != nil {
		return StatusResponse{}, err
	}
	completed := rec.State.CurrentStepID == "approved"
	return buildStatus(app, in, completed), nil
}

func (s *KYCService) SubmitProfile(ctx context.Context, runID string, p Profile) (StatusResponse, error) {
	if err := s.applicants.SaveProfile(runID, p); err != nil {
		return StatusResponse{}, err
	}
	return s.submitOnStep(ctx, runID, "collect-profile", map[string]any{
		"fullName": p.FullName,
		"email":    p.Email,
	})
}

func (s *KYCService) SubmitDocument(ctx context.Context, runID string, d Document) (StatusResponse, error) {
	if err := s.applicants.AddDocument(runID, d); err != nil {
		return StatusResponse{}, err
	}
	if err := FakeVendorCheckLiveness(runID); err != nil {
		return StatusResponse{}, err
	}
	needsReview := d.Type == "passport"
	return s.submitOnStep(ctx, runID, "document-upload", map[string]any{
		"documentType": d.Type,
		"documentRef":  d.Ref,
		"review": map[string]any{
			"required": needsReview,
		},
	})
}

func (s *KYCService) SubmitReview(ctx context.Context, runID string, approved bool) (StatusResponse, error) {
	return s.submitOnStep(ctx, runID, "manual-review", map[string]any{
		"approved": approved,
		"review": map[string]any{
			"approved": approved,
		},
	})
}

func (s *KYCService) submitOnStep(ctx context.Context, runID, wantStep string, input map[string]any) (StatusResponse, error) {
	in, err := restoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	if in.CurrentStepID() != wantStep {
		return StatusResponse{}, fmt.Errorf("wrong step: want %q, at %q", wantStep, in.CurrentStepID())
	}

	sub := in.SubmitInput(input)
	switch sub.Status {
	case engine.SubmitAdvanced, engine.SubmitStayOnStep:
		// ok
	case engine.SubmitFailed:
		return StatusResponse{}, fmt.Errorf("submit input: %w", sub.Err)
	default:
		return StatusResponse{}, fmt.Errorf("submit input: unexpected status %v", sub.Status)
	}

	if err := s.driveUntilBlocked(in); err != nil {
		return StatusResponse{}, err
	}
	if err := saveRun(ctx, s.runs, runID, in, s.def); err != nil {
		return StatusResponse{}, err
	}

	app, err := s.applicants.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	return buildStatus(app, in, in.IsTerminal()), nil
}

// driveUntilBlocked runs the engine until the next human pause or terminal step.
func (s *KYCService) driveUntilBlocked(in *engine.Instance) error {
	res := in.RunUntilBlocked()
	switch res.Status {
	case engine.RunBlocked, engine.RunCompleted:
		return nil
	case engine.RunFailed:
		return fmt.Errorf("run: %w", res.Err)
	default:
		return fmt.Errorf("run: unexpected status %v", res.Status)
	}
}
