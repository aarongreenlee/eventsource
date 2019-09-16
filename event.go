package eventsource

import (
	"time"
)

// Event implementations structure individual state changes to a resource.
//
// Naming Conventions
//
// An Event has happened and is immutable. When naming Events use past tense.
// Consider the following workflow:
//	* A user asks to change a name and issues the `ChangeName` command.
// 	* The system validates and accepts the command producing a `ChangedName` event.
type Event interface {
	// AggregateID implementations should return the id of the resource
	// an aggregation of events produces.
	AggregateID() string

	// EventVersion implementations should returns the resources current
	// version number.
	EventVersion() int64

	// EventAt implementations should return the timestamp marked when the
	// event occurred
	EventAt() time.Time

	// EventType returns the name of the event which is expected to be
	// past tense (e.g., "changedName")
	EventType() string
}
