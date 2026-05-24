package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/db"
)

type incomingRideRequestRepository interface {
	GetIncomingRequests(ctx context.Context, driverID string) ([]db.RideRequest, error)
}

type rideRequestActionRepository interface {
	GetRideRequestByID(ctx context.Context, id string) (db.RideRequest, error)
	SetRideRequestStatus(ctx context.Context, id, status string) error
}

type incomingRequestResponse struct {
	ID             string `json:"id"`
	RiderID        string `json:"rider_id"`
	RiderFirstName string `json:"rider_first_name"`
	RiderLastName  string `json:"rider_last_name"`
	PickupAddress  string `json:"pickup_address"`
	DropoffAddress string `json:"dropoff_address"`
	RequestedAt    string `json:"requested_at"`
	ExpiresAt      string `json:"expires_at"`
}

func GetIncomingRequests(repo incomingRideRequestRepository) http.HandlerFunc {
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

		requests, err := repo.GetIncomingRequests(r.Context(), sub)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not fetch incoming requests")
			return
		}

		resp := make([]incomingRequestResponse, len(requests))
		for i, rr := range requests {
			resp[i] = incomingRequestResponse{
				ID:             rr.ID,
				RiderID:        rr.RiderID,
				RiderFirstName: rr.RiderFirstName,
				RiderLastName:  rr.RiderLastName,
				PickupAddress:  rr.PickupAddress,
				DropoffAddress: rr.DropoffAddress,
				RequestedAt:    rr.RequestedAt.Format(time.RFC3339),
				ExpiresAt:      rr.ExpiresAt.Format(time.RFC3339),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func AcceptRideRequest(repo rideRequestActionRepository) http.HandlerFunc {
	return rideRequestAction(repo, "accepted", "request already accepted")
}

func DeclineRideRequest(repo rideRequestActionRepository) http.HandlerFunc {
	return rideRequestAction(repo, "declined", "request already declined")
}

func rideRequestAction(repo rideRequestActionRepository, newStatus, alreadyMsg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			writeError(w, http.StatusBadRequest, "sub is required")
			return
		}

		rr, err := repo.GetRideRequestByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrRideRequestNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not fetch ride request")
			return
		}

		if rr.DriverID != sub {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if rr.Status == "expired" {
			writeError(w, http.StatusGone, "request has expired")
			return
		}
		if rr.Status != "pending" {
			writeError(w, http.StatusConflict, alreadyMsg)
			return
		}

		if err := repo.SetRideRequestStatus(r.Context(), id, newStatus); err != nil {
			writeError(w, http.StatusInternalServerError, "could not update ride request")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	}
}
