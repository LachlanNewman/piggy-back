package handlers

import (
	"sync"

	"backend/events"
)

// recordingPublisher captures published events so tests can assert on who was
// notified and why.
type recordingPublisher struct {
	mu        sync.Mutex
	published []publishedEvent
}

type publishedEvent struct {
	Sub   string
	Event events.Event
}

func (p *recordingPublisher) Publish(sub string, e events.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.published = append(p.published, publishedEvent{Sub: sub, Event: e})
}

func (p *recordingPublisher) events() []publishedEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]publishedEvent(nil), p.published...)
}
