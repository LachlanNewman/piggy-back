package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	pool    *pgxpool.Pool
}

type location struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

type RideRequestMessage struct {
	RequestID   string   `json:"request_id"`
	RiderID     string   `json:"rider_id"`
	Pickup      location `json:"pickup"`
	Dropoff     location `json:"dropoff"`
	RequestedAt string   `json:"requested_at"`
}

func (m *RideRequestMessage) validate() error {
	if m.RequestID == "" {
		return fmt.Errorf("missing request_id")
	}
	if m.RiderID == "" {
		return fmt.Errorf("missing rider_id")
	}
	if m.Pickup.Address == "" || m.Pickup.Lat == 0 && m.Pickup.Lng == 0 {
		return fmt.Errorf("missing or zero pickup location")
	}
	if m.Dropoff.Address == "" || m.Dropoff.Lat == 0 && m.Dropoff.Lng == 0 {
		return fmt.Errorf("missing or zero dropoff location")
	}
	if m.RequestedAt == "" {
		return fmt.Errorf("missing requested_at")
	}
	return nil
}

func New(amqpURL string, pool *pgxpool.Pool) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare("ride-requests", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Consumer{conn: conn, channel: ch, pool: pool}, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	deliveries, err := c.channel.Consume("ride-requests", "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			c.handle(ctx, d)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, d amqp.Delivery) {
	var msg RideRequestMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("nack (no requeue): invalid JSON: %v", err)
		d.Nack(false, false)
		return
	}

	if err := msg.validate(); err != nil {
		log.Printf("nack (no requeue): validation failed: %v", err)
		d.Nack(false, false)
		return
	}

	requestedAt, err := time.Parse(time.RFC3339, msg.RequestedAt)
	if err != nil {
		log.Printf("nack (no requeue): invalid requested_at: %v", err)
		d.Nack(false, false)
		return
	}

	_, err = c.pool.Exec(ctx, `
		INSERT INTO ride_requests
			(request_id, rider_id, pickup_lat, pickup_lng, pickup_address,
			 dropoff_lat, dropoff_lng, dropoff_address, requested_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (request_id) DO NOTHING`,
		msg.RequestID,
		msg.RiderID,
		msg.Pickup.Lat, msg.Pickup.Lng, msg.Pickup.Address,
		msg.Dropoff.Lat, msg.Dropoff.Lng, msg.Dropoff.Address,
		requestedAt,
	)
	if err != nil {
		log.Printf("nack (requeue): db insert failed: %v", err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
