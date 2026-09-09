package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"backend/db"
	"backend/events"
)

// stubRepo is a minimal in-memory ride request store, enough to drive the
// create -> notify -> accept -> notify round trip over real HTTP.
type stubRepo struct {
	mu       sync.Mutex
	requests map[string]db.RideRequest
	nextID   int
}

func newStubRepo() *stubRepo {
	return &stubRepo{requests: make(map[string]db.RideRequest)}
}

func (s *stubRepo) CreateRideRequest(_ context.Context, p db.CreateRideRequestParams) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	id := "req-" + string(rune('a'+s.nextID-1))
	s.requests[id] = db.RideRequest{
		ID:             id,
		RiderID:        p.RiderID,
		DriverID:       p.DriverID,
		Status:         "pending",
		PickupAddress:  p.PickupAddress,
		DropoffAddress: p.DropoffAddress,
		RequestedAt:    time.Now(),
		ExpiresAt:      p.ExpiresAt,
	}
	return id, nil
}

func (s *stubRepo) GetRideRequestByID(_ context.Context, id string) (db.RideRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rr, ok := s.requests[id]
	if !ok {
		return db.RideRequest{}, db.ErrRideRequestNotFound
	}
	return rr, nil
}

func (s *stubRepo) HasActivePendingRequest(context.Context, string) (bool, error) { return false, nil }

func (s *stubRepo) SetRideRequestStatus(_ context.Context, id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rr, ok := s.requests[id]
	if !ok {
		return db.ErrRideRequestNotFound
	}
	rr.Status = status
	s.requests[id] = rr
	return nil
}

// readEvents reads SSE data frames off the stream and posts them to a channel.
func readEvents(t *testing.T, body *bufio.Reader) <-chan events.Event {
	t.Helper()
	out := make(chan events.Event, 8)
	go func() {
		defer close(out)
		for {
			line, err := body.ReadString('\n')
			if err != nil {
				return
			}
			payload, ok := strings.CutPrefix(strings.TrimSpace(line), "data: ")
			if !ok {
				continue // comment frame, retry hint, or blank separator
			}
			var e events.Event
			if err := json.Unmarshal([]byte(payload), &e); err != nil {
				return
			}
			out <- e
		}
	}()
	return out
}

func awaitEvent(t *testing.T, ch <-chan events.Event, want string) events.Event {
	t.Helper()
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				t.Fatalf("stream closed while waiting for %q", want)
			}
			if e.Type == want {
				return e
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for %q", want)
		}
	}
}

// TestRealtime_RideRequestRoundTrip drives the full live path over real HTTP:
// a carrier opens a stream, a rider posts a request, the carrier is told
// immediately, accepts, and the rider is told immediately.
func TestRealtime_RideRequestRoundTrip(t *testing.T) {
	repo := newStubRepo()
	hub := events.NewHub()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/events", Events(hub, time.Minute))
	mux.HandleFunc("/api/v1/ride-requests", CreateRideRequest(repo, 15*time.Minute, hub))
	mux.HandleFunc("/api/v1/ride-requests/{id}/accept", AcceptRideRequest(repo, hub))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	const carrier = "auth0|carrier"
	const rider = "auth0|rider"

	// Both parties open their live streams.
	carrierStream, closeCarrier := openStream(t, srv.URL, carrier)
	defer closeCarrier()
	riderStream, closeRider := openStream(t, srv.URL, rider)
	defer closeRider()

	awaitEvent(t, carrierStream, events.TypeConnected)
	awaitEvent(t, riderStream, events.TypeConnected)

	// Wait until the hub has actually registered both streams, so the publish
	// below cannot race the subscription.
	waitForSubscribers(t, hub, carrier, 1)
	waitForSubscribers(t, hub, rider, 1)

	// The rider sends a request.
	body := strings.NewReader(`{"pickup_address":"123 Main St","dropoff_address":"456 Oak Ave","driver_id":"` + carrier + `"}`)
	resp, err := http.Post(srv.URL+"/api/v1/ride-requests?sub="+rider, "application/json", body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create returned %d", resp.StatusCode)
	}

	// The carrier hears about it without polling.
	created := awaitEvent(t, carrierStream, events.TypeRideRequestCreated)
	if created.ID == "" {
		t.Fatal("created event carried no request id")
	}

	// ...and the rider is not told about their own request.
	select {
	case e := <-riderStream:
		t.Fatalf("rider received an event meant for the carrier: %+v", e)
	case <-time.After(100 * time.Millisecond):
	}

	// The carrier accepts.
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/ride-requests/"+created.ID+"/accept?sub="+carrier, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("accept returned %d", resp.StatusCode)
	}

	// The rider hears the answer immediately.
	updated := awaitEvent(t, riderStream, events.TypeRideRequestUpdated)
	if updated.ID != created.ID {
		t.Errorf("updated event id = %q, want %q", updated.ID, created.ID)
	}
	if updated.Status != "accepted" {
		t.Errorf("updated event status = %q, want accepted", updated.Status)
	}
}

func TestRealtime_StreamReleasedOnDisconnect(t *testing.T) {
	hub := events.NewHub()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/events", Events(hub, time.Minute))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	stream, closeStream := openStream(t, srv.URL, "auth0|alice")
	awaitEvent(t, stream, events.TypeConnected)
	waitForSubscribers(t, hub, "auth0|alice", 1)

	closeStream()

	// The handler must notice the dropped client and release its subscription,
	// otherwise every reconnect would leak one.
	waitForSubscribers(t, hub, "auth0|alice", 0)
}

func openStream(t *testing.T, baseURL, sub string) (<-chan events.Event, func()) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/events?sub="+sub, nil)
	if err != nil {
		t.Fatalf("build stream request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("stream Content-Type = %q", got)
	}
	return readEvents(t, bufio.NewReader(resp.Body)), func() { resp.Body.Close() }
}

func waitForSubscribers(t *testing.T, hub *events.Hub, sub string, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.Subscribers(sub) == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("want %d subscribers for %s, got %d", want, sub, hub.Subscribers(sub))
}
