package gob

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"sync"

	es "github.com/aarongreenlee/eventsource"
)

type gobEvent struct {
	Type string
	Data es.Event
}

// Serializer de/serializes events which have been bound to the serializer.
type Serializer struct {
	eventTypes map[string]struct{}
	m          sync.Mutex
}

// Bind registers the specified events with the serializer.
// Bind may be called multiple times.
func (s *Serializer) Bind(events ...es.Event) error {
	s.m.Lock()
	defer s.m.Unlock()

	// Allow calls to Bind to establish the event registry.
	if s.eventTypes == nil {
		s.eventTypes = make(map[string]struct{}, len(events))
	}

	for _, event := range events {
		gob.Register(event)
		eventType := event.EventType()

		if eventType == "" {
			return errors.New("unable to determine event type")
		}

		s.eventTypes[eventType] = struct{}{}
	}

	return nil
}

// MarshalEvent marshals the event into a Record which can be stored.
func (s *Serializer) MarshalEvent(v es.Event) (es.Record, error) {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(gobEvent{
		Type: v.EventType(),
		Data: v,
	})

	if err != nil {
		return es.Record{}, fmt.Errorf("unable to encode event: %w", err)
	}

	return es.Record{
		Version: v.EventVersion(),
		Data:    buffer.Bytes(),
	}, nil
}

// UnmarshalEvent converts the persistent type, Record, into an Event instance
func (s *Serializer) UnmarshalEvent(record es.Record) (es.Event, error) {
	event := gobEvent{}

	err := gob.NewDecoder(bytes.NewReader(record.Data)).Decode(&event)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal event: %w", err)
	}

	_, ok := s.eventTypes[event.Type]
	if !ok {
		return nil, fmt.Errorf("unbound event type: %q", event.Type)
	}

	// Sanity check for Event type casting.
	eventData, ok := event.Data.(es.Event)
	if !ok {
		return nil, fmt.Errorf("unable to cast to Event due to unknown data type %q", reflect.TypeOf(event.Data).Name())
	}

	return eventData, nil
}

// MarshalAll is a utility that marshals all the events provided into a History object
func (s *Serializer) MarshalAll(events ...es.Event) (es.History, error) {
	history := make(es.History, 0, len(events))

	for _, event := range events {
		record, err := s.MarshalEvent(event)
		if err != nil {
			return nil, err
		}
		history = append(history, record)
	}

	return history, nil
}

// New constructs a new gob serializer and populates it with the
// specified events. Bind may be subsequently called to add more events.
func New(events ...es.Event) (*Serializer, error) {
	serializer := &Serializer{
		eventTypes: make(map[string]struct{}),
	}

	if err := serializer.Bind(events...); err != nil {
		return nil, fmt.Errorf("failed to bind events to serializer: %w", err)
	}

	return serializer, nil
}
