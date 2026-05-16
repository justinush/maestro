package engine

import (
	"fmt"
	"sort"
	"strings"
)

// EventType identifies the kind of trace Event.
type EventType string

const (
	EventStepEntered     EventType = "step.entered"
	EventInputAccepted   EventType = "input.accepted"
	EventActionRan       EventType = "action.ran"
	EventTransitionGuard EventType = "transition.guard"
	EventTransitionTaken EventType = "transition.taken"
	EventBlocked         EventType = "run.blocked"
	EventCompleted       EventType = "run.completed"
)

// Event is one row in the execution trace (JSON-serializable).
type Event struct {
	Seq   int       `json:"seq"`
	RunID string    `json:"runId,omitempty"`
	Type  EventType `json:"type"`

	StepID string `json:"stepId,omitempty"`

	// input.accepted
	InputKeys []string `json:"inputKeys,omitempty"`

	// action.ran
	ActionList string `json:"actionList,omitempty"` // "onEnter" or "onExit"
	ActionID   string `json:"actionId,omitempty"`
	ActionType string `json:"actionType,omitempty"`

	// transition.guard / transition.taken:
	TransitionIndex int    `json:"transitionIndex,omitempty"`
	FromStepID      string `json:"fromStepId,omitempty"`
	ToStepID        string `json:"toStepId,omitempty"`

	// transition.guard:
	GuardResult string `json:"guardResult,omitempty"` // "true", "false", "error"
	GuardError  string `json:"guardError,omitempty"`
}

// String returns a short human-readable log line for demos and CLI output.
func (e Event) String() string {
	switch e.Type {
	case EventStepEntered:
		return fmt.Sprintf("%04d %s step=%q", e.Seq, e.Type, e.StepID)

	case EventInputAccepted:
		return fmt.Sprintf("%04d %s step=%q keys=%s", e.Seq, e.Type, e.StepID, fmtKeys(e.InputKeys))

	case EventActionRan:
		return fmt.Sprintf(
			"%04d %s step=%q list=%s action=%q type=%q",
			e.Seq, e.Type, e.StepID, e.ActionList, e.ActionID, e.ActionType,
		)

	case EventTransitionGuard:
		if e.GuardResult == "error" {
			return fmt.Sprintf(
				"%04d %s idx=%d from=%q to=%q result=%s error=%q",
				e.Seq, e.Type, e.TransitionIndex, e.FromStepID, e.ToStepID, e.GuardResult, e.GuardError,
			)
		}
		return fmt.Sprintf(
			"%04d %s idx=%d from=%q to=%q result=%s",
			e.Seq, e.Type, e.TransitionIndex, e.FromStepID, e.ToStepID, e.GuardResult,
		)

	case EventTransitionTaken:
		return fmt.Sprintf(
			"%04d %s idx=%d from=%q to=%q",
			e.Seq, e.Type, e.TransitionIndex, e.FromStepID, e.ToStepID,
		)

	case EventBlocked:
		return fmt.Sprintf("%04d %s step=%q", e.Seq, e.Type, e.StepID)

	case EventCompleted:
		return fmt.Sprintf("%04d %s step=%q", e.Seq, e.Type, e.StepID)

	default:
		return fmt.Sprintf("%04d %s step=%q", e.Seq, e.Type, e.StepID)
	}
}

func fmtKeys(keys []string) string {
	if len(keys) == 0 {
		return "[]"
	}
	cp := append([]string(nil), keys...)
	sort.Strings(cp)

	var b strings.Builder
	b.WriteByte('[')
	for i, k := range cp {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%q", k)
	}
	b.WriteByte(']')
	return b.String()
}
