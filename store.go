package eventsource

import (
	"context"
)

// Store provides an abstraction for the Repository to save data
type Store interface {
	// Save implementations should persist Record(s) to the store.
	Save(ctx context.Context, aggregateID string, records ...Record) error

	// Load implementations should read the records stored up to the
	// specified version or all events if toVersion is `0`. Specify the
	// fromVersion `0` to begin reading the tail of the event chain.
	Load(ctx context.Context, aggregateID string, fromVersion, toVersion int64) (History, error)
}

// Record is a serialized event suitable for storage.
type Record struct {
	Data    []byte
	Version int64
}

// History is a chain of events for a specific resource. Left-folding over
// these events will produce a resource aggregate.
type History []Record

// Len implements the sort.Interface.
func (h History) Len() int {
	return len(h)
}

// Swap implements the sort.Interface.
func (h History) Swap(a, b int) {
	h[a], h[b] = h[b], h[a]
}

// Less implements sort.Interface.
func (h History) Less(a, b int) bool {
	return h[a].Version < h[b].Version
}
