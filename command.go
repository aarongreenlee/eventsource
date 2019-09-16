package eventsource

import "context"

// Command encapsulates the data to mutate an aggregate
type Command interface {
	// AggregateID represents the id of the aggregate to apply to.
	AggregateID() string
	EventType() string
}

// CommandModel provides an embeddable struct that implements Command.
type CommandModel struct {
	ID   string
	Type string
}

// AggregateID implements the Command interface; returns the aggregate id
func (m CommandModel) AggregateID() string {
	return m.ID
}

func (m CommandModel) EventType() string {
	return m.Type
}

// CommandHandler consumes a command and emits Events
type CommandHandler interface {
	// Apply applies a command to an aggregate to generate a new set of events
	Apply(ctx context.Context, command Command) ([]Event, error)
}
