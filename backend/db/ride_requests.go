package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrRideRequestNotFound = errors.New("ride request not found")

type RideRequest struct {
	ID              string
	DriverID        string
	RiderID         string
	RiderFirstName  string
	RiderLastName   string
	Status          string
	PickupAddress   string
	DropoffAddress  string
	RequestedAt     time.Time
	ExpiresAt       time.Time
}

type CreateRideRequestParams struct {
	RiderID        string
	DriverID       string
	PickupAddress  string
	DropoffAddress string
	ExpiresAt      time.Time
}

func (d *DB) CreateRideRequest(ctx context.Context, p CreateRideRequestParams) (string, error) {
	var id string
	err := d.pool.QueryRow(ctx, `
		INSERT INTO ride_requests (request_id, rider_id, driver_id, pickup_address, dropoff_address, requested_at, expires_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, now(), $5)
		RETURNING id
	`, p.RiderID, p.DriverID, p.PickupAddress, p.DropoffAddress, p.ExpiresAt).Scan(&id)
	return id, err
}

func (d *DB) GetRideRequestByID(ctx context.Context, id string) (RideRequest, error) {
	var r RideRequest
	err := d.pool.QueryRow(ctx, `
		SELECT id, driver_id, rider_id, status, pickup_address, dropoff_address, requested_at, expires_at
		FROM ride_requests
		WHERE id = $1
	`, id).Scan(&r.ID, &r.DriverID, &r.RiderID, &r.Status, &r.PickupAddress, &r.DropoffAddress, &r.RequestedAt, &r.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RideRequest{}, ErrRideRequestNotFound
		}
		return RideRequest{}, err
	}
	if r.Status == "pending" && r.ExpiresAt.Before(time.Now()) {
		r.Status = "expired"
	}
	return r, nil
}

// HasActivePendingRequest returns true if the rider already has a non-expired pending request.
func (d *DB) HasActivePendingRequest(ctx context.Context, riderID string) (bool, error) {
	var exists bool
	err := d.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM ride_requests
			WHERE rider_id = $1 AND status = 'pending' AND expires_at > now()
		)
	`, riderID).Scan(&exists)
	return exists, err
}

type incomingRequestJSON struct {
	ID             string    `json:"id"`
	RiderID        string    `json:"rider_id"`
	RiderFirstName string    `json:"rider_first_name"`
	RiderLastName  string    `json:"rider_last_name"`
	PickupAddress  string    `json:"pickup_address"`
	DropoffAddress string    `json:"dropoff_address"`
	RequestedAt    time.Time `json:"requested_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// GetIncomingRequests returns all pending non-expired requests directed at driverID.
func (d *DB) GetIncomingRequests(ctx context.Context, driverID string) ([]RideRequest, error) {
	var raw json.RawMessage
	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(
			json_agg(json_build_object(
				'id',              rr.id::text,
				'rider_id',        rr.rider_id,
				'rider_first_name', COALESCE(u.first_name, ''),
				'rider_last_name',  COALESCE(u.last_name, ''),
				'pickup_address',  rr.pickup_address,
				'dropoff_address', rr.dropoff_address,
				'requested_at',    rr.requested_at,
				'expires_at',      rr.expires_at
			)),
			'[]'::json
		)
		FROM ride_requests rr
		LEFT JOIN users u ON u.auth_subject = rr.rider_id
		WHERE rr.driver_id = $1
		  AND rr.status = 'pending'
		  AND rr.expires_at > now()
	`, driverID).Scan(&raw)
	if err != nil {
		return nil, err
	}

	var rows []incomingRequestJSON
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	result := make([]RideRequest, len(rows))
	for i, r := range rows {
		result[i] = RideRequest{
			ID:             r.ID,
			RiderID:        r.RiderID,
			RiderFirstName: r.RiderFirstName,
			RiderLastName:  r.RiderLastName,
			PickupAddress:  r.PickupAddress,
			DropoffAddress: r.DropoffAddress,
			RequestedAt:    r.RequestedAt,
			ExpiresAt:      r.ExpiresAt,
		}
	}
	return result, nil
}

// SetRideRequestStatus updates the status of a ride request. Returns ErrRideRequestNotFound if not found.
func (d *DB) SetRideRequestStatus(ctx context.Context, id, status string) error {
	var discard string
	err := d.pool.QueryRow(ctx, `
		UPDATE ride_requests SET status = $1 WHERE id = $2 RETURNING id
	`, status, id).Scan(&discard)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRideRequestNotFound
		}
		return err
	}
	return nil
}
