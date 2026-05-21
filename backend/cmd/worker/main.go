package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/consumer"
	"backend/db"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.New(ctx)
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer pool.Close()

	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		log.Fatal("RABBITMQ_URL not set")
	}

	c, err := consumer.New(amqpURL, pool)
	if err != nil {
		log.Fatalf("consumer init: %v", err)
	}
	defer c.Close()

	done := make(chan error, 1)
	go func() {
		log.Println("worker started, consuming ride-requests")
		done <- c.Consume(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("received signal %s, shutting down", sig)
		cancel()
		<-done
	case err := <-done:
		if err != nil {
			log.Fatalf("consumer error: %v", err)
		}
	}
}
