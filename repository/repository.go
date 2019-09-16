package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	es "github.com/aarongreenlee/eventsource"
	"github.com/aarongreenlee/eventsource/serializer/gob"
	"github.com/aarongreenlee/eventsource/store/memory"
)

// Repository provides the primary abstraction to saving and loading events
type Repository struct {
	debug      bool
	observers  []func(es.Event)
	prototype  reflect.Type
	serializer es.Serializer
	store      es.Store
	writer     io.Writer
}

// Option provides functional configuration for a *Repository
type Option func(*Repository) error

// WithStore may be used to configure a new store.
func WithStore(store es.Store) Option {
	return func(r *Repository) error {
		if store == nil {
			return errors.New("must not provided a nil store")
		}
		r.store = store
		return nil
	}
}

// WithSerializer specifies the serializer to be used.
func WithSerializer(serializer es.Serializer) Option {
	return func(r *Repository) error {
		r.serializer = serializer
		return nil
	}
}

// WithEvents binds one or more events to the Serializer. If you choose not to
// use the default serializer you need to provide the WithSerializer option
// before this option or the events will be bound to the default serializer
// instead of the one you later specify.
func WithEvents(events ...es.Event) Option {
	return func(r *Repository) error {
		if r.serializer == nil {
			return errors.New("a serializer must have been configured for the repository before binding events")
		}
		return r.serializer.Bind(events...)
	}
}

// WithObservers allows observers to watch the saved events; Observers should
// invoke very short lived operations as calls will block until the observer is
// finished
func WithObservers(observers ...func(event es.Event)) Option {
	return func(r *Repository) error {
		r.observers = append(r.observers, observers...)
		return nil
	}
}

// New creates a new Repository using defaults and then applying any
// operations.
//
// Defaults
//
// A repository is built with the following defaults:
//
//	* Memory store
//	* Gob serializer
func New(prototype es.Aggregate, events []es.Event, opts ...Option) (*Repository, error) {
	t := reflect.TypeOf(prototype)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	r := &Repository{prototype: t}

	defaultSerializer, err := gob.New()
	if err != nil {
		return nil, fmt.Errorf("error building default serializer: %w", err)
	}

	defaults := []Option{
		WithStore(memory.New()),
		WithSerializer(defaultSerializer),
	}

	for _, opt := range defaults {
		err := opt(r)
		if err != nil {
			return nil, fmt.Errorf("error applying default configuration: %w", err)
		}
	}

	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	if err := WithEvents(events...)(r); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) logf(format string, args ...interface{}) {
	if !r.debug {
		return
	}

	now := time.Now().UTC().Format(time.StampMilli)
	_, _ = io.WriteString(r.writer, now)
	_, _ = io.WriteString(r.writer, " ")

	_, _ = fmt.Fprintf(r.writer, format, args...)
	if !strings.HasSuffix(format, "\n") {
		_, _ = io.WriteString(r.writer, "\n")
	}
}

// New returns a new instance of the aggregate
func (r *Repository) New() es.Aggregate {
	return reflect.New(r.prototype).Interface().(es.Aggregate)
}

// Save persists the events into the underlying Store
func (r *Repository) Save(ctx context.Context, events ...es.Event) error {
	if len(events) == 0 {
		return nil
	}

	aggregateID := events[0].AggregateID()

	history := make(es.History, 0, len(events))
	for _, event := range events {
		record, err := r.serializer.MarshalEvent(event)
		if err != nil {
			return err
		}

		history = append(history, record)
	}

	return r.store.Save(ctx, aggregateID, history...)
}

// Load retrieves the specified aggregate from the underlying store
func (r *Repository) Load(ctx context.Context, aggregateID string) (es.Aggregate, error) {
	v, _, err := r.loadVersion(ctx, aggregateID)
	return v, err
}

// loadVersion loads the specified aggregate from the store and returns both the Aggregate and the
// current version number of the aggregate
func (r *Repository) loadVersion(ctx context.Context, aggregateID string) (es.Aggregate, int64, error) {
	history, err := r.store.Load(ctx, aggregateID, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	entryCount := len(history)
	if entryCount == 0 {
		return nil, 0, fmt.Errorf("unable to load %v, %s", r.New(), aggregateID)
	}

	r.logf("Loaded %d event(s) for aggregate id, %s", entryCount, aggregateID)
	aggregate := r.New()

	var version int64

	for _, record := range history {
		event, err := r.serializer.UnmarshalEvent(record)
		if err != nil {
			return nil, 0, err
		}

		err = aggregate.On(event)
		if err != nil {
			eventType := event.EventType()
			return nil, 0, fmt.Errorf("repository for %q aggregate was unable to handle event, %v: this is a programming error which may be solved by updating the On function of the repository: error %s", r.prototype.Name(), eventType, err)
		}

		version = event.EventVersion()
	}

	return aggregate, version, nil
}

// Apply executes the command specified and returns the current version of the
// aggregate
func (r *Repository) Apply(ctx context.Context, command es.Command) (int64, error) {
	if command == nil {
		return 0, errors.New("command provided to Repository.Apply must not be nil")
	}

	aggregateID := command.AggregateID()
	if aggregateID == "" {
		return 0, errors.New("command provided to Repository.Apply must not contain a blank AggregateID")
	}

	aggregate, version, err := r.loadVersion(ctx, aggregateID)

	if err != nil {
		aggregate = r.New()
	}

	h, ok := aggregate.(es.CommandHandler)
	if !ok {
		return 0, fmt.Errorf("aggregate, %T, does not implement CommandHandler", aggregate)
	}

	events, err := h.Apply(ctx, command)
	if err != nil {
		return 0, err
	}

	if len(events) == 0 {
		return -1, es.ErrNoEventsProduced
	}

	err = r.Save(ctx, events...)
	if err != nil {
		return 0, err
	}
	version = events[len(events)-1].EventVersion()

	// publish events to observers
	if r.observers != nil {
		for _, event := range events {
			for _, observer := range r.observers {
				observer(event)
			}
		}
	}

	return version, nil
}

// Store returns the underlying Store
func (r *Repository) Store() es.Store {
	return r.store
}

// Serializer returns the underlying serializer
func (r *Repository) Serializer() es.Serializer {
	return r.serializer
}
