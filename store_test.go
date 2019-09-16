package eventsource_test

import (
	"sort"
	"testing"

	"github.com/aarongreenlee/eventsource"
)

// TestHistorySort asserts that the chain of events which establishes
// history can be sorted by Version.
func TestHistorySort(t *testing.T) {
	history := eventsource.History{
		{Version: 1},
		{Version: 3},
		{Version: 4},
		{Version: 2},
	}

	sort.Sort(history)

	for i := 0; i < len(history); i++ {
		expect := int64(i + 1)
		if expect != history[i].Version {
			t.Fatalf("expected version %d but found %d", expect, history[i].Version)
		}
	}
}
