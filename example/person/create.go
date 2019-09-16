package person

import (
	"context"
	"fmt"
	"time"

	"github.com/aarongreenlee/eventsource"
	"github.com/aarongreenlee/eventsource/example/audit"
	"github.com/aarongreenlee/eventsource/example/idgen"
	"github.com/aarongreenlee/eventsource/example/session"
)

const (
	// CreateCommandKey names the command which can be issued to request that
	// a Person be created.
	CreateCommandKey = "createPerson"
	// CreatedEventKey names the immutable event stored in history.
	CreatedEventKey = "createdPerson"
)

const (
	ErrRecreateAttempt = Error("the provided Person resource has a version greater than 0")
	ErrStateExists     = Error("the provided Person has one or more populated attributes")
)

// CreateRequest establishes a contract for a client issuing a command to
// create a new person resource.
type CreateRequest struct {
	Name  string
	Email string
}

// CreateResponse establishes the contract for a response to a successful
// create response.
type CreateResponse struct {
	Person Person
}

// HandleCreate event handler for requests to create a person.
func (s Service) Create(ctx context.Context, req CreateRequest) (CreateResponse, error) {

	cmd := CreateCommand{
		CommandModel: eventsource.CommandModel{
			ID:   idgen.NewID(),
			Type: CreateCommandKey,
		},
		Data: CreateEvent{
			Name:  req.Name,
			Email: req.Email,
			Audit: audit.Create{
				Created:     time.Now(),
				CreatedBy:   session.Username(ctx),
				CreatedByID: session.UserID(ctx),
			},
		},
	}

	version, err := s.repo.Apply(ctx, cmd)
	if err != nil {
		return CreateResponse{}, fmt.Errorf("error creating Person: %w", err)
	}

	return CreateResponse{
		Person: Person{
			ID:      cmd.CommandModel.ID,
			Email:   cmd.Data.Email,
			Name:    cmd.Data.Name,
			Version: version,
		},
	}, nil
}

// CreateCommand structures the command to change the system state by
// creating a new Person.
type CreateCommand struct {
	eventsource.CommandModel
	Data CreateEvent
}

// CreateEvent is the event data produced by a accepted CreateCommand
// and is serialized and persisted by the eventsource.Repository.
type CreateEvent struct {
	ID      string
	Version int64

	Name  string
	Email string

	Audit audit.Create
}

// AggregateID implements the eventsource.Event interface.
func (e CreateEvent) AggregateID() string {
	return e.ID
}

// EventVersion implements the eventsource.Event interface.
func (e CreateEvent) EventVersion() int64 {
	return e.Version
}

// EventAt implements the eventsource.Event interface.
func (e CreateEvent) EventAt() time.Time {
	return e.Audit.Created
}

// EventType implements the eventsource.Event interface.
func (e CreateEvent) EventType() string {
	return CreatedEventKey
}

// apply applies the command to create a person after first verifying that
// the change can be applied to the person provided.
func (cmd *CreateCommand) apply(p *Person) ([]eventsource.Event, error) {

	// Verify that the current state of Person is acceptable to apply
	// this create command.
	switch {
	case p.Version != 0:
		return nil, ErrRecreateAttempt
	case p.ID != "", p.Email != "", p.Name != "":
		return nil, ErrStateExists
	}

	// Verify the user data provided when issuing the command is valid.
	if err := p.ValidateEmail(cmd.Data.Email); err != nil {
		return nil, err
	}
	if err := p.ValidateName(cmd.Data.Name); err != nil {
		return nil, err
	}

	// We also want to validate data not provided by the user. Although we
	// built the Audit event ourselves we want to be sure we're happy with
	// the command data before producing the event. Once the event is
	// produced we'll have built our immutable record.
	cmd.Data.Audit.Event = CreatedEventKey
	if err := cmd.Data.Audit.Validate(); err != nil {
		return nil, err
	}

	// Having passed all validations we are ready to produce our event.
	// At this point, our language will change from "create" to "created".
	event := &CreateEvent{
		Name:    cmd.Data.Name,
		Email:   cmd.Data.Email,
		Audit:   cmd.Data.Audit,
		Version: 1,
	}

	return []eventsource.Event{event}, nil
}
