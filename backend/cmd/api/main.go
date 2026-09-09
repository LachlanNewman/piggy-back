package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"backend/auth"
	"backend/config"
	"backend/db"
	"backend/handlers"

	"github.com/caarlos0/env/v11"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		ms := time.Since(start).Milliseconds()
		if rec.status >= 400 && rec.status < 500 {
			slog.Warn("bad request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", ms, "response", rec.body.String())
		} else {
			slog.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", ms)
		}
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v", err)
	}

	authn, err := auth.New(ctx, cfg.OIDCIssuer, cfg.OIDCAudience)
	if err != nil {
		log.Fatalf("auth init: %v", err)
	}

	pool, err := db.New(ctx)
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello from Go!"})
	})

	userDB := db.NewDB(pool)

	ttl := time.Duration(cfg.RideRequestTTLMinutes) * time.Minute
	staleThreshold := time.Duration(cfg.LocationPollIntervalSecs*2) * time.Second

	// Every /api/v1 route carries user data, so all of them require a verified
	// token. /api/health and /api/hello stay public for Render's health check.
	protected := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, authn.Middleware(h))
	}

	protected("/api/v1/users", handlers.CreateUser(userDB))
	protected("/api/v1/users/me", handlers.GetUserMe(userDB))
	protected("/api/v1/users/nearby", handlers.GetNearbyUsers(userDB, cfg.NearbyRadiusKm, staleThreshold))

	protected("/api/v1/location", handlers.PushLocation(userDB))

	protected("/api/v1/ride-requests", handlers.CreateRideRequest(userDB, ttl))
	protected("/api/v1/ride-requests/incoming", handlers.GetIncomingRequests(userDB))
	protected("/api/v1/ride-requests/{id}", handlers.GetRideRequest(userDB))
	protected("/api/v1/ride-requests/{id}/accept", handlers.AcceptRideRequest(userDB))
	protected("/api/v1/ride-requests/{id}/decline", handlers.DeclineRideRequest(userDB))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("backend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, requestLogger(cors(mux))))
}
