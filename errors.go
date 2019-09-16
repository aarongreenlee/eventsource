package eventsource

const (
	// ErrNotFound should be returned by store implementations when a
	// aggregate could not be found.
	ErrNotFound = Error("not found")

	// ErrNoEventsProduced is returned when changes are applied to produce a
	// new aggregate but the operation results in no new events being
	// produced. Such a condition may not be unexpected depending on the
	// context.
	ErrNoEventsProduced = Error("no events produced")
)

// Type Error implements the Error interface and is allows for errors to be
// exported as constants.
type Error string

// Error implements the standard go Error interface.
func (e Error) Error() string {
	return string(e)
}
