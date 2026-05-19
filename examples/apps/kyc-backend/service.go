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

// afterWorkflowSuccess persists app-owned data only after Maestro accepted input and the run was saved.
type afterWorkflowSuccess func() error

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

	s.applicants.Create(applicantID, runID)

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

func (s *KYCService) GetEvents(ctx context.Context, runID string) (EventsResponse, error) {
	if _, err := s.applicants.GetByRunID(runID); err != nil {
		return EventsResponse{}, err
	}
	in, err := restoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return EventsResponse{}, err
	}
	events := in.Events()
	lines := make([]string, len(events))
	for i := range events {
		lines[i] = events[i].String()
	}
	return EventsResponse{RunID: runID, Events: lines}, nil
}

func (s *KYCService) SubmitProfile(ctx context.Context, runID string, p Profile) (StatusResponse, error) {
	if err := p.Validate(); err != nil {
		return StatusResponse{}, err
	}
	return s.submitOnStep(ctx, runID, "collect-profile", map[string]any{
		"fullName": p.FullName,
		"email":    p.Email,
	}, func() error {
		return s.applicants.SaveProfile(runID, p)
	})
}

func (s *KYCService) SubmitDocument(ctx context.Context, runID string, d Document) (StatusResponse, error) {
	if err := d.Validate(); err != nil {
		return StatusResponse{}, err
	}
	app, err := s.applicants.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	if err := FakeVendorCheckLiveness(app.ApplicantID); err != nil {
		return StatusResponse{}, err
	}
	needsReview := d.Type == "passport"
	return s.submitOnStep(ctx, runID, "document-upload", map[string]any{
		"documentType": d.Type,
		"documentRef":  d.Ref,
		"review": map[string]any{
			"required": needsReview,
		},
	}, func() error {
		return s.applicants.AddDocument(runID, d)
	})
}

func (s *KYCService) SubmitReview(ctx context.Context, runID string, approved bool) (StatusResponse, error) {
	return s.submitOnStep(ctx, runID, "manual-review", map[string]any{
		"approved": approved,
		"review": map[string]any{
			"approved": approved,
		},
	}, nil)
}

func (s *KYCService) submitOnStep(
	ctx context.Context,
	runID, wantStep string,
	input map[string]any,
	after afterWorkflowSuccess,
) (StatusResponse, error) {
	in, err := restoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	if in.CurrentStepID() != wantStep {
		return StatusResponse{}, fmt.Errorf("%w: want %q, at %q", ErrWrongStep, wantStep, in.CurrentStepID())
	}

	sub := in.SubmitInput(input)
	switch sub.Status {
	case engine.SubmitAdvanced, engine.SubmitStayOnStep:
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

	if after != nil {
		if err := after(); err != nil {
			return StatusResponse{}, err
		}
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
