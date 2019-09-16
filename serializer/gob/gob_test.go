package gob_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	es "github.com/aarongreenlee/eventsource"
	"github.com/aarongreenlee/eventsource/serializer/gob"

	"github.com/stretchr/testify/assert"
)

const eventAType = "eventA"

type EventA struct {
	Name string `json:"name"`
	Event
}

const eventBType = "eventB"

type EventB struct {
	Description string
	Event
}

func TestGobSerializer(t *testing.T) {
	type testCase struct {
		event  es.Event
		record es.Record
	}

	testCases := []testCase{{
		event: EventA{
			Event: Event{
				ID:      "1",
				Version: 1,
				Type:    eventAType,
			},
			Name: "Alpha",
		},
	}, {
		event: EventB{
			Event: Event{
				ID:      "2",
				Version: 2,
				Type:    eventAType,
			},
			Description: "An event which is tested",
		},
	}, {
		event: EventA{
			Event: Event{
				ID:      "3",
				Version: 3,
				Type:    eventAType,
			},
			Name: "Beta",
		},
	}, {
		event: EventB{
			Event: Event{
				ID:      "4",
				Version: 4,
				Type:    eventBType,
			},
			Description: "An event which is tested again.",
		},
	}}

	serializer, err := gob.New(
		&EventA{
			Event: Event{
				Type: eventAType,
			},
		},
		&EventB{
			Event: Event{
				Type: eventBType,
			},
		},
	)

	require.NoError(t, err)

	t.Run("TestMarshalEvent", func(t *testing.T) {
		for i, tc := range testCases {
			record, err := serializer.MarshalEvent(tc.event)
			assert.Nil(t, err)
			testCases[i].record = record // make available for next test
		}
	})

	t.Run("TestUnmarshalEvent", func(t *testing.T) {
		for _, tc := range testCases {
			v, err := serializer.UnmarshalEvent(tc.record)
			assert.Nil(t, err)
			switch v.(type) {
			case *EventA:
				event := tc.event.(EventA)
				assert.Equal(t, &event, v.(*EventA))
			case *EventB:
				event := tc.event.(EventB)
				assert.Equal(t, &event, v.(*EventB))
			default:
				assert.Fail(t, "programming error in test: the event type %T is unsupported by the test", v)
			}
		}
	})
}

// TestMarshalAll asserts the MarshalAll function can de/serialize.
func TestMarshalAll(t *testing.T) {
	event := EventA{
		Event: Event{
			ID:      "ABC",
			Version: 786,
			Type:    eventAType,
		},
		Name: "Alpha Omega Beta",
	}

	serializer, err := gob.New(&event)
	require.NoError(t, err)

	history, err := serializer.MarshalAll(event)
	assert.Nil(t, err)
	assert.NotNil(t, history)

	v, err := serializer.UnmarshalEvent(history[0])
	assert.Nil(t, err)

	found, ok := v.(*EventA)
	assert.True(t, ok)
	assert.Equal(t, &event, found)
}

// Event implements eventsource.Event interface for our test cases.
type Event struct {
	// ID contains the AggregateID
	ID string `json:"ID"`

	// Version contains the EventVersion
	Version int64 `json:"version"`

	// At contains the EventAt
	At time.Time `json:"eventAt"`

	// Type identifies the type of event
	Type string `json:"type"`
}

func (e Event) AggregateID() string {
	return e.ID
}

func (e Event) EventVersion() int64 {
	return e.Version
}

func (e Event) EventAt() time.Time {
	return e.At
}

func (e Event) EventType() string {
	return e.Type
}
