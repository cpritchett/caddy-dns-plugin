package docker

import (
	"testing"
	"time"
)

func TestBuildEventFiltersIncludesSwarmTypes(t *testing.T) {
	filters := buildEventFilters(Options{IncludeSwarm: true})

	if !contains(filters.Types, EventTypeContainer) {
		t.Fatalf("expected container type in filters, got %v", filters.Types)
	}
	if !contains(filters.Types, EventTypeService) {
		t.Fatalf("expected service type in filters, got %v", filters.Types)
	}

	for _, action := range []string{"create", "start", "stop", "die", "destroy", "remove", "update"} {
		if !contains(filters.Actions, action) {
			t.Fatalf("expected %q action filter, got %v", action, filters.Actions)
		}
	}
}

func TestWatcherDebounce(t *testing.T) {
	watcher := NewWatcher(nil, Options{Debounce: 2 * time.Second})
	anchor := time.Date(2026, 2, 2, 12, 0, 0, 0, time.UTC)

	first := Event{ID: "abc", Type: EventTypeContainer, Time: anchor}
	if !watcher.shouldEmit(first) {
		t.Fatal("expected first event to emit")
	}

	second := Event{ID: "abc", Type: EventTypeContainer, Time: anchor.Add(1 * time.Second)}
	if watcher.shouldEmit(second) {
		t.Fatal("expected second event within debounce window to be suppressed")
	}

	third := Event{ID: "abc", Type: EventTypeContainer, Time: anchor.Add(3 * time.Second)}
	if !watcher.shouldEmit(third) {
		t.Fatal("expected event after debounce window to emit")
	}

	other := Event{ID: "xyz", Type: EventTypeContainer, Time: anchor.Add(1 * time.Second)}
	if !watcher.shouldEmit(other) {
		t.Fatal("expected different container to emit regardless of debounce")
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
