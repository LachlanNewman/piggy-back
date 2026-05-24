package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/db"
)

const maxNearbyRadiusKm = 20.0

type nearbyUserRepository interface {
	GetUserLocation(ctx context.Context, sub string) (lat, lng float64, err error)
	GetNearbyUsers(ctx context.Context, sub string, lat, lng, radiusKm float64, staleThreshold time.Duration) ([]db.NearbyUser, error)
}

type nearbyUserResponse struct {
	ID          int    `json:"id"`
	AuthSubject string `json:"auth_subject"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

func GetNearbyUsers(repo nearbyUserRepository, radiusKm float64, staleThreshold time.Duration) http.HandlerFunc {
	if radiusKm > maxNearbyRadiusKm {
		radiusKm = maxNearbyRadiusKm
	}

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

		lat, lng, err := repo.GetUserLocation(r.Context(), sub)
		if err != nil {
			if errors.Is(err, db.ErrLocationNotFound) {
				writeError(w, http.StatusNotFound, "location not found — push your location first")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not fetch location")
			return
		}

		nearby, err := repo.GetNearbyUsers(r.Context(), sub, lat, lng, radiusKm, staleThreshold)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not fetch nearby users")
			return
		}

		resp := make([]nearbyUserResponse, len(nearby))
		for i, u := range nearby {
			resp[i] = nearbyUserResponse{ID: u.ID, AuthSubject: u.AuthSubject, FirstName: u.FirstName, LastName: u.LastName}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
