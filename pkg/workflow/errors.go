package workflow

import "errors"

var (
	// ErrNotFound is returned when no workflow is registered for the given key.
	ErrNotFound = errors.New("workflow: not found")

	// ErrDuplicateKey is returned when Register would overwrite an existing key.
	ErrDuplicateKey = errors.New("workflow: duplicate key")

	// ErrDefinitionMismatch is returned when a RunRecord's workflow identity does not
	// match the loaded definition (RestoreInstance).
	ErrDefinitionMismatch = errors.New("workflow: definition mismatch")
)
