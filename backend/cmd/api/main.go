package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"backend/config"
	"backend/db"
	"backend/handlers"

	"github.com/caarlos0/env/v11"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
	mux.HandleFunc("/api/v1/users", handlers.CreateUser(userDB))
	mux.HandleFunc("/api/v1/users/me", handlers.GetUserMe(userDB))

	log.Println("backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", cors(mux)))
}
