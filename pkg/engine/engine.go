// Package engine is the stable low-level API for executing workflow definitions.
//
// Most applications should start with [github.com/justinush/maestro/pkg/maestro]:
//
//	maestro.Load -> Runtime.NewInstance -> RunUntilBlocked -> SubmitInput
//
// Use pkg/engine when you need direct control over [Options], action registries,
// or [Snapshot] restore without the maestro.Runtime facade.
//
// Typical control flow: branch on [RunResult].Status and [SubmitInputResult].Status.
// Err is set only when status is [RunFailed] or [SubmitFailed].
//
//	NewInstance
//	    -> RunUntilBlocked
//	    |-- RunCompleted -> done
//	    |-- RunFailed (inspect Err) -> done
//	    |-- RunBlocked
//	            -> SubmitInput
//	            |-- SubmitFailed (inspect Err) -> done
//	            |-- SubmitAdvanced or SubmitStayOnStep
//	                    -> RunUntilBlocked (repeat)
package engine

import (
	"errors"
	"net/http"

	iengine "github.com/justinush/maestro/internal/engine"
	"github.com/justinush/maestro/pkg/definition"
)

// NewInstance starts execution at def.InitialStepID. def must be non-nil.
func NewInstance(def *definition.WorkflowDefinition, opts Options) (*Instance, error) {
	in, err := iengine.NewInstance(def, opts.toInternal())
	if err != nil {
		return nil, err
	}
	return &Instance{Instance: in}, nil
}

// NewInstanceFromSnapshot restores execution state from snap using the same workflow definition as before.
func NewInstanceFromSnapshot(def *definition.WorkflowDefinition, snap Snapshot, opts Options) (*Instance, error) {
	in, err := iengine.NewInstanceFromSnapshot(def, snap, opts.toInternal())
	if err != nil {
		return nil, err
	}
	return &Instance{Instance: in}, nil
}

// NewRegistry creates an empty action registry.
func NewRegistry() *Registry {
	return &Registry{Registry: iengine.NewRegistry()}
}

// DefaultRegistry includes the built-in "stub" runner.
func DefaultRegistry() *Registry {
	return &Registry{Registry: iengine.DefaultRegistry()}
}

// RegistryWithHTTP returns DefaultRegistry plus an "http" runner using client.
// Pass your own *http.Client in production; the maestro simulate CLI uses a private long-timeout client.
func RegistryWithHTTP(client *http.Client) *Registry {
	return &Registry{Registry: iengine.RegistryWithHTTP(client)}
}

// NewStubRunner returns the built-in stub ActionRunner (usually accessed via DefaultRegistry).
func NewStubRunner() ActionRunner {
	return iengine.NewStubRunner()
}

// NewHTTPRunner returns an ActionRunner for "http" actions.
func NewHTTPRunner(client *http.Client) ActionRunner {
	return iengine.NewHTTPRunner(client)
}

// AsInputValidationError reports whether err unwraps to *InputValidationError (e.g. SubmitInputResult.Err).
func AsInputValidationError(err error) (*InputValidationError, bool) {
	if v, ok := errors.AsType[*InputValidationError](err); ok {
		return v, true
	}
	return nil, false
}

// AsUnknownStepError reports whether err unwraps to *UnknownStepError (e.g. from Instance.StepByID).
func AsUnknownStepError(err error) (*UnknownStepError, bool) {
	if v, ok := errors.AsType[*UnknownStepError](err); ok {
		return v, true
	}
	return nil, false
}
