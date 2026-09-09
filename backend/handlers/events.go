package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"backend/events"
)

// eventSubscriber opens a per-user event stream. Implemented by *events.Hub.
type eventSubscriber interface {
	Subscribe(sub string) (<-chan events.Event, func())
}

// eventPublisher emits a change signal to a user's open streams.
// Implemented by *events.Hub.
type eventPublisher interface {
	Publish(sub string, e events.Event)
}

// clientRetryMS tells EventSource how long to wait before reconnecting after
// the stream drops. Browsers default to ~3s; a shorter delay gets a rider back
// on the wire fast without hammering the server.
const clientRetryMS = 2000

// Events streams change signals for one user over Server-Sent Events.
//
// SSE rather than WebSockets: traffic here is one-directional (every client
// action is already a REST call), EventSource reconnects on its own, and it
// needs nothing outside net/http.
func Events(hub eventSubscriber, heartbeat time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			writeError(w, http.StatusBadRequest, "sub is required")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			slog.Error("event stream needs a flushable ResponseWriter")
			writeError(w, http.StatusInternalServerError, "streaming unsupported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		// no-transform stops intermediaries compressing the stream. A gzipping
		// proxy buffers until its window fills, which strands events for
		// minutes — the exact failure this endpoint exists to avoid.
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		// Belt and braces for nginx, which buffers proxied responses by default.
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		stream, release := hub.Subscribe(sub)
		defer release()

		fmt.Fprintf(w, "retry: %d\n\n", clientRetryMS)
		writeSSE(w, events.Event{Type: events.TypeConnected})
		flusher.Flush()

		slog.Info("event stream opened", "sub", sub)
		defer slog.Info("event stream closed", "sub", sub)

		ticker := time.NewTicker(heartbeat)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case e := <-stream:
				writeSSE(w, e)
				flusher.Flush()
			case <-ticker.C:
				// A comment frame keeps idle connections alive through proxies
				// and load balancers that would otherwise time them out.
				fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, e events.Event) {
	payload, err := json.Marshal(e)
	if err != nil {
		slog.Error("marshal event failed", "type", e.Type, "err", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", payload)
}
