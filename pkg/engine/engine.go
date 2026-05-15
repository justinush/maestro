package engine

import (
	"errors"
	"net/http"

	iengine "github.com/justinush/maestro/internal/engine"
	"github.com/justinush/maestro/pkg/definition"
)

type (
	Instance             = iengine.Instance
	Options              = iengine.Options
	Registry             = iengine.Registry
	ActionContext        = iengine.ActionContext
	ActionRunner         = iengine.ActionRunner
	Event                = iengine.Event
	EventType            = iengine.EventType
	InputValidationError = iengine.InputValidationError
	UnknownStepError     = iengine.UnknownStepError
	Snapshot             = iengine.Snapshot
	RunStatus            = iengine.RunStatus
	RunResult            = iengine.RunResult
	SubmitInputStatus    = iengine.SubmitInputStatus
	SubmitInputResult    = iengine.SubmitInputResult
)

const (
	EventStepEntered     EventType = iengine.EventStepEntered
	EventInputAccepted   EventType = iengine.EventInputAccepted
	EventActionRan       EventType = iengine.EventActionRan
	EventTransitionGuard EventType = iengine.EventTransitionGuard
	EventTransitionTaken EventType = iengine.EventTransitionTaken
	EventBlocked         EventType = iengine.EventBlocked
	EventCompleted       EventType = iengine.EventCompleted
)

const (
	RunBlocked   RunStatus = iengine.RunBlocked
	RunCompleted RunStatus = iengine.RunCompleted
	RunFailed    RunStatus = iengine.RunFailed
)

const (
	SubmitStayOnStep SubmitInputStatus = iengine.SubmitStayOnStep
	SubmitAdvanced   SubmitInputStatus = iengine.SubmitAdvanced
	SubmitFailed     SubmitInputStatus = iengine.SubmitFailed
)

var (
	ErrNilDefinition        = iengine.ErrNilDefinition
	ErrEmptyInitialStepID   = iengine.ErrEmptyInitialStepID
	ErrInitialStepUnknown   = iengine.ErrInitialStepUnknown
	ErrEmptyStepID          = iengine.ErrEmptyStepID
	ErrEmptyStepKind        = iengine.ErrEmptyStepKind
	ErrDuplicateStepID      = iengine.ErrDuplicateStepID
	ErrNoMatchingTransition = iengine.ErrNoMatchingTransition
	ErrUnknownActionType    = iengine.ErrUnknownActionType
	ErrCELGuard             = iengine.ErrCELGuard
)

// NewInstance starts a run at the workflow initial step. def must be non-nil.
func NewInstance(def *definition.WorkflowDefinition, opts Options) (*Instance, error) {
	return iengine.NewInstance(def, opts)
}

// NewInstanceFromSnapshot restores state from a snapshot.
// Use the same workflow definition (id, version, graph) as when the snapshot was taken.
func NewInstanceFromSnapshot(def *definition.WorkflowDefinition, snap Snapshot, opts Options) (*Instance, error) {
	return iengine.NewInstanceFromSnapshot(def, snap, opts)
}

func NewRegistry() *Registry {
	return iengine.NewRegistry()
}

func DefaultRegistry() *Registry {
	return iengine.DefaultRegistry()
}

func RegistryWithHTTP(client *http.Client) *Registry {
	return iengine.RegistryWithHTTP(client)
}

func SimulateHTTPClient() *http.Client {
	return iengine.SimulateHTTPClient()
}

func NewStubRunner() ActionRunner {
	return iengine.NewStubRunner()
}

func NewHTTPRunner(client *http.Client) ActionRunner {
	return iengine.NewHTTPRunner(client)
}

// AsInputValidationError reports whether err is an *InputValidationError from SubmitInput.
func AsInputValidationError(err error) (*InputValidationError, bool) {
	if v, ok := errors.AsType[*InputValidationError](err); ok {
		return v, true
	}
	return nil, false
}

// AsUnknownStepError reports whether err is an *UnknownStepError.
func AsUnknownStepError(err error) (*UnknownStepError, bool) {
	if v, ok := errors.AsType[*UnknownStepError](err); ok {
		return v, true
	}
	return nil, false
}
