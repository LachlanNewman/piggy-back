package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrRideRequestNotFound = errors.New("ride request not found")

type RideRequest struct {
	ID             string
	Status         string
	PickupAddress  string
	DropoffAddress string
	RequestedAt    time.Time
}

type CreateRideRequestParams struct {
	RiderID        string
	PickupAddress  string
	DropoffAddress string
}

func (d *DB) CreateRideRequest(ctx context.Context, p CreateRideRequestParams) (string, error) {
	var id string
	err := d.pool.QueryRow(ctx, `
		INSERT INTO ride_requests (request_id, rider_id, pickup_address, dropoff_address, requested_at)
		VALUES (gen_random_uuid(), $1, $2, $3, now())
		RETURNING id
	`, p.RiderID, p.PickupAddress, p.DropoffAddress).Scan(&id)
	return id, err
}

func (d *DB) GetRideRequestByID(ctx context.Context, id string) (RideRequest, error) {
	var r RideRequest
	err := d.pool.QueryRow(ctx, `
		SELECT id, status, pickup_address, dropoff_address, requested_at
		FROM ride_requests
		WHERE id = $1
	`, id).Scan(&r.ID, &r.Status, &r.PickupAddress, &r.DropoffAddress, &r.RequestedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RideRequest{}, ErrRideRequestNotFound
		}
		return RideRequest{}, err
	}
	return r, nil
}
