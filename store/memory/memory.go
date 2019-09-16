// Package memory implements a in-memory story which supports the
// Event Source package which is intended to be used for test cases.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	es "github.com/aarongreenlee/eventsource"
)

// memoryStore provides an in-memory implementation of Store
type memoryStore struct {
	*sync.Mutex
	eventsByID map[string]es.History
}

// New produces a new memory store that meets the eventsource.Store interface.
func New() *memoryStore {
	return &memoryStore{
		Mutex:      &sync.Mutex{},
		eventsByID: map[string]es.History{},
	}
}

// Save stores the record(s) in memory which then become history.
func (m *memoryStore) Save(ctx context.Context, aggregateID string, records ...es.Record) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.eventsByID[aggregateID]; !ok {
		m.eventsByID[aggregateID] = es.History{}
	}

	history := append(m.eventsByID[aggregateID], records...)
	sort.Sort(history)
	m.eventsByID[aggregateID] = history

	return nil
}

// Load returns the history from memory if any.
func (m *memoryStore) Load(ctx context.Context, aggregateID string, fromVersion, toVersion int64) (es.History, error) {
	m.Lock()
	defer m.Unlock()

	all, ok := m.eventsByID[aggregateID]
	if !ok {

		for id := range m.eventsByID {
			fmt.Printf("asked for %q, found %q\n", aggregateID, id)
		}

		return nil, es.ErrNotFound
	}

	history := make(es.History, 0, len(all))
	if len(all) > 0 {
		for _, record := range all {
			if v := record.Version; v >= fromVersion && (toVersion == 0 || v <= toVersion) {
				history = append(history, record)
			}
		}
	}

	return all, nil
}
