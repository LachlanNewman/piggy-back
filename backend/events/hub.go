// Package events provides an in-process publish/subscribe hub used to push
// live changes to connected clients over Server-Sent Events.
//
// The hub is deliberately in-process: it fans out to the subscribers held by
// this backend instance only. Running more than one backend replica would need
// a shared broker (Postgres LISTEN/NOTIFY is the natural fit here) behind the
// same Publish/Subscribe surface.
package events

import "sync"

// Event types pushed to clients.
const (
	// TypeConnected is sent once when a stream opens, so the client can show
	// that it is live without waiting for real traffic.
	TypeConnected = "connected"
	// TypeRideRequestCreated tells a carrier a new request is addressed to them.
	TypeRideRequestCreated = "ride_request.created"
	// TypeRideRequestUpdated tells a rider their request changed status.
	TypeRideRequestUpdated = "ride_request.updated"
)

// Event is a change signal, not a data carrier: it names what changed so the
// client can refetch the authoritative representation from the REST API. That
// keeps one source of truth for response shapes.
type Event struct {
	Type   string `json:"type"`
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}

// buffer is the per-subscriber queue depth. Events are tiny change signals and
// the client resyncs on receipt, so a short queue is plenty.
const buffer = 16

// Hub fans events out to subscribers, keyed by the user's auth subject.
// It is safe for concurrent use.
type Hub struct {
	mu   sync.Mutex
	subs map[string]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[string]map[chan Event]struct{})}
}

// Subscribe opens a stream of events addressed to sub. The returned release
// func must be called when the stream ends; it is safe to call more than once.
// A user may hold several subscriptions at once (multiple tabs or devices).
func (h *Hub) Subscribe(sub string) (<-chan Event, func()) {
	ch := make(chan Event, buffer)

	h.mu.Lock()
	if h.subs[sub] == nil {
		h.subs[sub] = make(map[chan Event]struct{})
	}
	h.subs[sub][ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	release := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			set, ok := h.subs[sub]
			if !ok {
				return
			}
			if _, ok := set[ch]; ok {
				delete(set, ch)
				close(ch)
			}
			if len(set) == 0 {
				delete(h.subs, sub)
			}
		})
	}
	return ch, release
}

// Publish delivers e to every stream open for sub.
//
// Delivery is best effort: a subscriber whose buffer is full is skipped rather
// than blocking the caller, so one stalled client can never hold up the HTTP
// handler that triggered the event. Such a client still converges — it refetches
// on its next event and on its fallback poll.
//
// The lock is held across the sends, which is what makes closing a channel in
// release safe: a send can never race with a close.
func (h *Hub) Publish(sub string, e Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs[sub] {
		select {
		case ch <- e:
		default:
		}
	}
}

// Subscribers reports how many streams are open for sub. Used by tests and
// logging.
func (h *Hub) Subscribers(sub string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs[sub])
}
