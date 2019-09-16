package person

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aarongreenlee/eventsource"
	"github.com/aarongreenlee/eventsource/example/audit"
)

// Person represents our domain resource for this example.
type Person struct {
	ID          string
	Name        string
	Email       string
	Version     int64
	CreateAudit audit.Create
}

// ValidateName verifies a name is not empty.
func (p Person) ValidateName(s string) error {
	if s == "" {
		return errors.New("name must not be empty")
	}
	return nil
}

// ValidateEmail is a very simple validation of an email value for
// this example application.
func (p Person) ValidateEmail(s string) error {
	switch {
	case !strings.Contains(s, "@"):
		fallthrough
	case !strings.Contains(s, "."):
		fallthrough
	case len(s) < 5: // must be at least 5 characters: "a@b.c"
		return fmt.Errorf("%w: email appears invalid", ErrFailedValidation)
	}

	return nil
}

// Apply is our write handler and implements the `eventsource.CommandHandler` interface and is called
// when we wish to execute a command. Each command is implemented individually.
//
// Each command should validate that the change can be applied to the current
// state. An accepted command results in one or more events being returned
// which will be persisted by the eventsource Repository.
func (p *Person) Apply(_ context.Context, cmd eventsource.Command) ([]eventsource.Event, error) {
	switch v := cmd.(type) {
	case CreateCommand:
		return v.apply(p)
	}

	return nil, fmt.Errorf("the command %q is not implemented on %T", cmd.EventType(), p)
}

// On is our read handler and implements the `eventsource.Aggregate` interface
// and is responsible for building the aggregate record. The eventsource
// repository will call this function once for each event in the resource's
// history where they are applied to produce the Aggregate. In functional terms,
// this is a left-fold over events when reading the record.
func (p *Person) On(event eventsource.Event) error {
	switch event.EventType() {
	case CreatedEventKey:
		e, ok := event.(CreateEvent)
		if !ok {
			return fmt.Errorf("could not cast for event %q", CreatedEventKey)
		}
		if err := e.on(p); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("the command %q is not implemented on %T", event.EventType(), p)
}
