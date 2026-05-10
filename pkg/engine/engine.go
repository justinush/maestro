package engine

import (
	"errors"
	"net/http"

	"github.com/justinush/maestro/internal/definition"
	iengine "github.com/justinush/maestro/internal/engine"
)

// Types
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
)

// Event kinds (trace)
const (
	EventStepEntered     EventType = iengine.EventStepEntered
	EventInputAccepted   EventType = iengine.EventInputAccepted
	EventActionRan       EventType = iengine.EventActionRan
	EventTransitionGuard EventType = iengine.EventTransitionGuard
	EventTransitionTaken EventType = iengine.EventTransitionTaken
	EventBlocked         EventType = iengine.EventBlocked
	EventCompleted       EventType = iengine.EventCompleted
)

// Sentinel errors
var (
	ErrNilDefinition        = iengine.ErrNilDefinition
	ErrEmptyInitialStepID   = iengine.ErrEmptyInitialStepID
	ErrInitialStepUnknown   = iengine.ErrInitialStepUnknown
	ErrEmptyStepID          = iengine.ErrEmptyStepID
	ErrEmptyStepKind        = iengine.ErrEmptyStepKind
	ErrDuplicateStepID      = iengine.ErrDuplicateStepID
	ErrNeedsInput           = iengine.ErrNeedsInput
	ErrWorkflowCompleted    = iengine.ErrWorkflowCompleted
	ErrNoMatchingTransition = iengine.ErrNoMatchingTransition
	ErrUnknownActionType    = iengine.ErrUnknownActionType
	ErrCELGuard             = iengine.ErrCELGuard
)

// NewInstance creates an instance at initialStepId. def must be non-nil.
func NewInstance(def *definition.WorkflowDefinition, opts Options) (*Instance, error) {
	return iengine.NewInstance(def, opts)
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

// AsInputValidationError returns (*InputValidationError, true) if err wraps that type.
func AsInputValidationError(err error) (*InputValidationError, bool) {
	if v, ok := errors.AsType[*InputValidationError](err); ok {
		return v, true
	}
	return nil, false
}

// AsUnknownStepError returns (*UnknownStepError, true) if err is that type.
func AsUnknownStepError(err error) (*UnknownStepError, bool) {
	if v, ok := errors.AsType[*UnknownStepError](err); ok {
		return v, true
	}
	return nil, false
}
