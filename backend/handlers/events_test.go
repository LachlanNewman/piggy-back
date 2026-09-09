package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/events"
)

// fakeHub hands the test direct control of the stream a handler reads from.
type fakeHub struct {
	ch         chan events.Event
	subscribed string
	released   chan struct{}
}

func newFakeHub() *fakeHub {
	return &fakeHub{ch: make(chan events.Event, 4), released: make(chan struct{}, 1)}
}

func (f *fakeHub) Subscribe(sub string) (<-chan events.Event, func()) {
	f.subscribed = sub
	return f.ch, func() { f.released <- struct{}{} }
}

func TestEvents_MissingSub(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	w := httptest.NewRecorder()

	Events(newFakeHub(), time.Minute).ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestEvents_MethodNotAllowed(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/events?sub=auth0|alice", nil)
	w := httptest.NewRecorder()

	Events(newFakeHub(), time.Minute).ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", w.Code)
	}
}

func TestEvents_SendsStreamHeadersAndHandshake(t *testing.T) {
	hub := newFakeHub()
	ctx, cancel := context.WithCancel(context.Background())
	r := httptest.NewRequest(http.MethodGet, "/api/v1/events?sub=auth0|alice", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		Events(hub, time.Minute).ServeHTTP(w, r)
		close(done)
	}()

	// The handshake frame is written before the handler blocks, so cancelling
	// immediately still leaves it in the recorder.
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if got := w.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q", got)
	}
	// no-transform is load-bearing: without it a compressing proxy buffers the
	// stream and events arrive minutes late, or not at all.
	if got := w.Header().Get("Cache-Control"); got != "no-cache, no-transform" {
		t.Errorf("Cache-Control = %q, want it to include no-transform", got)
	}
	if got := w.Header().Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q", got)
	}

	body := w.Body.String()
	if !strings.Contains(body, "retry: ") {
		t.Errorf("missing retry hint in %q", body)
	}
	if !strings.Contains(body, `"type":"connected"`) {
		t.Errorf("missing connected handshake in %q", body)
	}
	if hub.subscribed != "auth0|alice" {
		t.Errorf("subscribed as %q", hub.subscribed)
	}
}

func TestEvents_ForwardsPublishedEvent(t *testing.T) {
	hub := newFakeHub()
	ctx, cancel := context.WithCancel(context.Background())
	r := httptest.NewRequest(http.MethodGet, "/api/v1/events?sub=auth0|alice", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		Events(hub, time.Minute).ServeHTTP(w, r)
		close(done)
	}()

	hub.ch <- events.Event{Type: events.TypeRideRequestUpdated, ID: "req-1", Status: "accepted"}
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	body := w.Body.String()
	want := `data: {"type":"ride_request.updated","id":"req-1","status":"accepted"}`
	if !strings.Contains(body, want) {
		t.Fatalf("want %q in body, got %q", want, body)
	}
}

func TestEvents_HeartbeatKeepsConnectionWarm(t *testing.T) {
	hub := newFakeHub()
	ctx, cancel := context.WithCancel(context.Background())
	r := httptest.NewRequest(http.MethodGet, "/api/v1/events?sub=auth0|alice", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		Events(hub, 10*time.Millisecond).ServeHTTP(w, r)
		close(done)
	}()

	time.Sleep(80 * time.Millisecond)
	cancel()
	<-done

	if !strings.Contains(w.Body.String(), ": ping") {
		t.Fatalf("want a heartbeat comment, got %q", w.Body.String())
	}
}

func TestEvents_ReleasesSubscriptionWhenClientDisconnects(t *testing.T) {
	hub := newFakeHub()
	ctx, cancel := context.WithCancel(context.Background())
	r := httptest.NewRequest(http.MethodGet, "/api/v1/events?sub=auth0|alice", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		Events(hub, time.Minute).ServeHTTP(w, r)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	select {
	case <-hub.released:
	default:
		t.Fatal("subscription was not released on disconnect")
	}
}
