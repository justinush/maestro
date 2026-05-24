package engine

import iengine "github.com/justinush/maestro/internal/engine"

// ActionContext is the input to [ActionRunner.Run] for one action invocation.
type ActionContext = iengine.ActionContext

// ActionRunner executes a single workflow action (stub, http, or custom types).
type ActionRunner = iengine.ActionRunner

// EventType identifies the kind of trace [Event].
type EventType = iengine.EventType

// Event is one row in the execution trace (JSON-serializable).
type Event = iengine.Event

// EventType constants for [Instance.Events] and [RunResult.Events].
const (
	// EventStepEntered is recorded when execution enters a step.
	EventStepEntered EventType = iengine.EventStepEntered
	// EventInputAccepted is recorded when [Instance.SubmitInput] accepts input on a human step.
	EventInputAccepted EventType = iengine.EventInputAccepted
	// EventActionRan is recorded after a successful onEnter/onExit action.
	EventActionRan EventType = iengine.EventActionRan
	// EventTransitionGuard is recorded when TraceGuards is enabled and a guard is evaluated.
	EventTransitionGuard EventType = iengine.EventTransitionGuard
	// EventTransitionTaken is recorded when an outbound transition fires.
	EventTransitionTaken EventType = iengine.EventTransitionTaken
	// EventBlocked is recorded when [RunUntilBlocked] stops on a human step.
	EventBlocked EventType = iengine.EventBlocked
	// EventCompleted is recorded when execution reaches a terminal end step.
	EventCompleted EventType = iengine.EventCompleted
)

// InputValidationError means [Instance.SubmitInput] input failed JSON Schema validation for a human step.
type InputValidationError = iengine.InputValidationError

// UnknownStepError means a step id was not found in the workflow (see [Instance.StepByID]).
type UnknownStepError = iengine.UnknownStepError

// Snapshot is JSON-friendly runtime state: run id, step position, variables, and trace.
// Compiled CEL programs and input schemas are rebuilt after restore.
type Snapshot = iengine.Snapshot

// RunStatus explains why [Instance.RunUntilBlocked] stopped.
type RunStatus = iengine.RunStatus

// RunStatus values returned by [Instance.RunUntilBlocked].
const (
	// RunBlocked means execution paused on a human step; call [Instance.SubmitInput] next.
	RunBlocked RunStatus = iengine.RunBlocked
	// RunCompleted means execution reached a declared terminal end step.
	RunCompleted RunStatus = iengine.RunCompleted
	// RunFailed means execution stopped with an error; see [RunResult.Err].
	RunFailed RunStatus = iengine.RunFailed
)

// RunResult is the outcome of [Instance.RunUntilBlocked].
//
// Events is a point-in-time snapshot (matches [Instance.Events] at return).
// Err is non-nil only when Status is [RunFailed].
type RunResult = iengine.RunResult

// SubmitInputStatus is the outcome of [Instance.SubmitInput].
type SubmitInputStatus = iengine.SubmitInputStatus

// SubmitInputStatus values returned by [Instance.SubmitInput].
const (
	// SubmitStayOnStep means input was stored but no outbound transition matched.
	SubmitStayOnStep SubmitInputStatus = iengine.SubmitStayOnStep
	// SubmitAdvanced means the instance moved to SubmitInputResult.StepID.
	SubmitAdvanced SubmitInputStatus = iengine.SubmitAdvanced
	// SubmitFailed means validation or leaving the step failed; see [SubmitInputResult.Err].
	SubmitFailed SubmitInputStatus = iengine.SubmitFailed
)

// SubmitInputResult is returned by [Instance.SubmitInput].
//
// StepID is the current step after the call.
// Err is set only when Status is [SubmitFailed].
type SubmitInputResult = iengine.SubmitInputResult

// Errors returned while building or executing instances.
var (
	// ErrNilDefinition is returned when a nil workflow definition is passed to constructors.
	ErrNilDefinition = iengine.ErrNilDefinition
	// ErrEmptyInitialStepID is returned when initialStepId is empty.
	ErrEmptyInitialStepID = iengine.ErrEmptyInitialStepID
	// ErrInitialStepUnknown is returned when initialStepId is not present in steps.
	ErrInitialStepUnknown = iengine.ErrInitialStepUnknown
	// ErrEmptyStepID is returned when a step id is empty.
	ErrEmptyStepID = iengine.ErrEmptyStepID
	// ErrEmptyStepKind is returned when a step kind is empty.
	ErrEmptyStepKind = iengine.ErrEmptyStepKind
	// ErrDuplicateStepID is returned when two steps share the same id.
	ErrDuplicateStepID = iengine.ErrDuplicateStepID
	// ErrNoMatchingTransition is returned when every outbound guard evaluated to false.
	ErrNoMatchingTransition = iengine.ErrNoMatchingTransition
	// ErrUnknownActionType is returned when an action type is not registered.
	ErrUnknownActionType = iengine.ErrUnknownActionType
	// ErrCELGuard is returned when a transition guard fails to evaluate.
	ErrCELGuard = iengine.ErrCELGuard
)
