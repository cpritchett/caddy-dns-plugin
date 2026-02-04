package docker

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	EventTypeContainer = "container"
	EventTypeService   = "service"
)

var (
	errMissingSource = errors.New("docker watcher requires an event source")
)

type Event struct {
	ID         string
	Name       string
	Type       string
	Action     string
	Attributes map[string]string
	Time       time.Time
}

type Filters struct {
	Types   []string
	Actions []string
}

type EventSource interface {
	Events(ctx context.Context, filters Filters) (<-chan Event, <-chan error)
}

type Options struct {
	IncludeSwarm bool
	Debounce     time.Duration
}

type Watcher struct {
	source    EventSource
	opts      Options
	now       func() time.Time
	mu        sync.Mutex
	lastEvent map[string]time.Time
}

func NewWatcher(source EventSource, opts Options) *Watcher {
	if opts.Debounce < 0 {
		opts.Debounce = 0
	}

	return &Watcher{
		source:    source,
		opts:      opts,
		now:       time.Now,
		lastEvent: make(map[string]time.Time),
	}
}

func (w *Watcher) Run(ctx context.Context) (<-chan Event, <-chan error) {
	out := make(chan Event)
	errs := make(chan error, 1)

	go w.run(ctx, out, errs)

	return out, errs
}

func (w *Watcher) run(ctx context.Context, out chan<- Event, errs chan<- error) {
	defer close(out)
	defer close(errs)

	if w.source == nil {
		errs <- errMissingSource
		return
	}

	messages, errStream := w.source.Events(ctx, buildEventFilters(w.opts))

	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-errStream:
			if !ok {
				return
			}
			if err != nil {
				errs <- err
			}
			return
		case message, ok := <-messages:
			if !ok {
				return
			}

			if !w.shouldEmit(message) {
				continue
			}

			select {
			case out <- message:
			case <-ctx.Done():
				return
			}
		}
	}
}

func buildEventFilters(opts Options) Filters {
	filters := Filters{
		Types:   []string{EventTypeContainer},
		Actions: []string{"create", "start", "stop", "die", "destroy", "remove", "update"},
	}

	if opts.IncludeSwarm {
		filters.Types = append(filters.Types, EventTypeService)
	}

	return filters
}

func (w *Watcher) shouldEmit(event Event) bool {
	if w.opts.Debounce == 0 {
		return true
	}

	key := eventKey(event)
	if key == "" {
		return true
	}

	when := event.Time
	if when.IsZero() {
		when = w.now()
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if last, ok := w.lastEvent[key]; ok {
		if when.Sub(last) < w.opts.Debounce {
			return false
		}
	}

	w.lastEvent[key] = when
	return true
}

func eventKey(event Event) string {
	id := event.ID
	if id == "" {
		return ""
	}

	if event.Type == "" {
		return id
	}

	return event.Type + ":" + id
}
