package eventsource

// Aggregate represents the current state of the domain resource
// and can be thought of as a left fold over events.
type Aggregate interface {
	// On will be called for each event in the resource's history.
	// Implementations should produce an error if the event could not be
	// applied.
	On(event Event) error
}
