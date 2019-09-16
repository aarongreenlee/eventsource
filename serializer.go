package eventsource

// Serializer implementations should serialize Events so they can be stored.
// Once serialized, an Event is called a Record.
type Serializer interface {
	// Bind registers one or more events to the Serializer
	Bind(events ...Event) error

	// MarshalEvent implementations should serialize an Event into a Record
	// suitable for storage.
	MarshalEvent(event Event) (Record, error)

	// UnmarshalEvent implementations should deserialize a Record into an Event.
	UnmarshalEvent(record Record) (Event, error)
}
