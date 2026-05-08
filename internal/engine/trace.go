package engine

import (
	"fmt"
	"sort"
	"strings"
)

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

type Event struct {
	Seq  int
	Type EventType

	StepID string

	// For input.accepted:
	InputKeys []string

	// For action.ran:
	ActionList string // "onEnter" or "onExit"
	ActionID   string
	ActionType string

	// For transition.guard / transition.taken:
	TransitionIndex int
	FromStepID      string
	ToStepID        string

	// For transition.guard:
	GuardResult string // "true", "false", "error"
	GuardError  string // short error string when GuardResult == "error"
}

func (e Event) String() string {
	switch e.Type {
	case EventStepEntered:
		return fmt.Sprintf("%04d %s step=%q", e.Seq, e.Type, e.StepID)

	case EventInputAccepted:
		return fmt.Sprintf("%04d %s step=%q keys=%s", e.Seq, e.Type, e.StepID, fmtKeys(e.InputKeys))

	case EventActionRan:
		return fmt.Sprintf("%04d %s step=%q list=%s action=%q type=%q",
			e.Seq, e.Type, e.StepID, e.ActionList, e.ActionID, e.ActionType,
		)

	case EventTransitionGuard:
		if e.GuardResult == "error" {
			return fmt.Sprintf("%04d %s idx=%d from=%q to=%q result=%s error=%q",
				e.Seq, e.Type, e.TransitionIndex, e.FromStepID, e.ToStepID, e.GuardResult, e.GuardError,
			)
		}
		return fmt.Sprintf("%04d %s idx=%d from=%q to=%q result=%s",
			e.Seq, e.Type, e.TransitionIndex, e.FromStepID, e.ToStepID, e.GuardResult,
		)

	case EventTransitionTaken:
		return fmt.Sprintf("%04d %s idx=%d from=%q to=%q",
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
