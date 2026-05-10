package run

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// MemoryStore is a process-local Store for tests and demos (data lost on exit).
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*RunRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*RunRecord)}
}

func (m *MemoryStore) Create(ctx context.Context, rec *RunRecord) error {
	_ = ctx
	if rec == nil || rec.RunID == "" {
		return fmt.Errorf("run: invalid record")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[rec.RunID]; ok {
		return ErrExists
	}
	stored, err := cloneRecord(rec)
	if err != nil {
		return err
	}
	stored.Revision = 1
	m.data[rec.RunID] = stored
	return nil
}

func (m *MemoryStore) Get(ctx context.Context, runID string) (*RunRecord, error) {
	_ = ctx
	if runID == "" {
		return nil, ErrNotFound
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.data[runID]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneRecord(cur)
}

func (m *MemoryStore) Save(ctx context.Context, rec *RunRecord) error {
	_ = ctx
	if rec == nil || rec.RunID == "" {
		return ErrNotFound
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.data[rec.RunID]
	if !ok {
		return ErrNotFound
	}
	if rec.Revision != cur.Revision {
		return ErrRevisionConflict
	}
	next, err := cloneRecord(rec)
	if err != nil {
		return err
	}
	next.Revision = cur.Revision + 1
	m.data[rec.RunID] = next
	return nil
}

func cloneRecord(r *RunRecord) (*RunRecord, error) {
	if r == nil {
		return nil, nil
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var out RunRecord
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
