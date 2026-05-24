package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/db"
)

type rideRequestRepository interface {
	CreateRideRequest(ctx context.Context, p db.CreateRideRequestParams) (string, error)
	GetRideRequestByID(ctx context.Context, id string) (db.RideRequest, error)
	HasActivePendingRequest(ctx context.Context, riderID string) (bool, error)
}

type createRideRequestBody struct {
	PickupAddress  string `json:"pickup_address"`
	DropoffAddress string `json:"dropoff_address"`
	DriverID       string `json:"driver_id"`
}

type createRideRequestResponse struct {
	ID string `json:"id"`
}

type getRideRequestResponse struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	PickupAddress  string `json:"pickup_address"`
	DropoffAddress string `json:"dropoff_address"`
	RequestedAt    string `json:"requested_at"`
	ExpiresAt      string `json:"expires_at"`
}

func CreateRideRequest(repo rideRequestRepository, ttl time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			writeError(w, http.StatusBadRequest, "sub is required")
			return
		}

		var body createRideRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if body.PickupAddress == "" {
			writeError(w, http.StatusBadRequest, "pickup_address is required")
			return
		}
		if body.DropoffAddress == "" {
			writeError(w, http.StatusBadRequest, "dropoff_address is required")
			return
		}
		if body.DriverID == "" {
			writeError(w, http.StatusBadRequest, "driver_id is required")
			return
		}

		has, err := repo.HasActivePendingRequest(r.Context(), sub)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create ride request")
			return
		}
		if has {
			writeError(w, http.StatusConflict, "you already have an active request")
			return
		}

		id, err := repo.CreateRideRequest(r.Context(), db.CreateRideRequestParams{
			RiderID:        sub,
			DriverID:       body.DriverID,
			PickupAddress:  body.PickupAddress,
			DropoffAddress: body.DropoffAddress,
			ExpiresAt:      time.Now().Add(ttl),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create ride request")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createRideRequestResponse{ID: id})
	}
}

func GetRideRequest(repo rideRequestRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}

		rr, err := repo.GetRideRequestByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrRideRequestNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not get ride request")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getRideRequestResponse{
			ID:             rr.ID,
			Status:         rr.Status,
			PickupAddress:  rr.PickupAddress,
			DropoffAddress: rr.DropoffAddress,
			RequestedAt:    rr.RequestedAt.Format(time.RFC3339),
			ExpiresAt:      rr.ExpiresAt.Format(time.RFC3339),
		})
	}
}
