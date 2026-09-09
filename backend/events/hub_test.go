package events

import (
	"sync"
	"testing"
)

func TestHub_PublishReachesSubscriber(t *testing.T) {
	h := NewHub()
	ch, release := h.Subscribe("auth0|alice")
	defer release()

	h.Publish("auth0|alice", Event{Type: TypeRideRequestCreated, ID: "req-1"})

	got := <-ch
	if got.Type != TypeRideRequestCreated || got.ID != "req-1" {
		t.Fatalf("got %+v", got)
	}
}

func TestHub_PublishIsScopedToSubject(t *testing.T) {
	h := NewHub()
	alice, releaseAlice := h.Subscribe("auth0|alice")
	defer releaseAlice()
	bob, releaseBob := h.Subscribe("auth0|bob")
	defer releaseBob()

	h.Publish("auth0|bob", Event{Type: TypeRideRequestCreated, ID: "req-1"})

	select {
	case e := <-alice:
		t.Fatalf("alice received an event addressed to bob: %+v", e)
	default:
	}

	if got := <-bob; got.ID != "req-1" {
		t.Fatalf("bob got %+v", got)
	}
}

func TestHub_FansOutToEverySubscription(t *testing.T) {
	h := NewHub()
	tab1, release1 := h.Subscribe("auth0|alice")
	defer release1()
	tab2, release2 := h.Subscribe("auth0|alice")
	defer release2()

	if n := h.Subscribers("auth0|alice"); n != 2 {
		t.Fatalf("want 2 subscribers, got %d", n)
	}

	h.Publish("auth0|alice", Event{Type: TypeRideRequestUpdated, ID: "req-1", Status: "accepted"})

	for i, ch := range []<-chan Event{tab1, tab2} {
		got := <-ch
		if got.Status != "accepted" {
			t.Fatalf("subscription %d got %+v", i, got)
		}
	}
}

func TestHub_ReleaseStopsDelivery(t *testing.T) {
	h := NewHub()
	ch, release := h.Subscribe("auth0|alice")

	release()

	if n := h.Subscribers("auth0|alice"); n != 0 {
		t.Fatalf("want 0 subscribers after release, got %d", n)
	}

	// Publishing to a released subject must not panic on the closed channel.
	h.Publish("auth0|alice", Event{Type: TypeRideRequestCreated, ID: "req-1"})

	if _, open := <-ch; open {
		t.Fatal("channel should be closed after release")
	}
}

func TestHub_ReleaseIsIdempotent(t *testing.T) {
	h := NewHub()
	_, release := h.Subscribe("auth0|alice")

	release()
	release() // must not panic by double-closing
}

func TestHub_SlowSubscriberDoesNotBlockPublish(t *testing.T) {
	h := NewHub()
	_, release := h.Subscribe("auth0|alice")
	defer release()

	// Overflow the buffer many times over. Publish must stay non-blocking, so
	// this returns rather than deadlocking the test.
	for i := 0; i < buffer*10; i++ {
		h.Publish("auth0|alice", Event{Type: TypeRideRequestCreated, ID: "req"})
	}
}

func TestHub_ConcurrentSubscribeReleasePublish(t *testing.T) {
	h := NewHub()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, release := h.Subscribe("auth0|alice")
			release()
		}()
		go func() {
			defer wg.Done()
			h.Publish("auth0|alice", Event{Type: TypeRideRequestCreated, ID: "req"})
		}()
	}

	wg.Wait()
	if n := h.Subscribers("auth0|alice"); n != 0 {
		t.Fatalf("want 0 subscribers, got %d", n)
	}
}
