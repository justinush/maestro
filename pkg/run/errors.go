package run

import "errors"

var (
	// ErrNotFound is returned when no run exists for the given id.
	ErrNotFound = errors.New("run: not found")

	// ErrExists is returned when Create is called for an id that already exists.
	ErrExists = errors.New("run: already exists")

	// ErrRevisionConflict is returned when Save is called with a stale revision.
	ErrRevisionConflict = errors.New("run: revision conflict")
)
