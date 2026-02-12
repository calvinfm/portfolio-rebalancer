package main

import (
	"context"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/kafka"
	"portfolio-rebalancer/internal/storage"

	"github.com/joho/godotenv"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load("C:\\Go\\src\\portfolio-rebalancer\\.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	// Initializing elasticsearch if needed
	if err := storage.InitElastic(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	// Initializing kafka if needed
	if err := kafka.InitKafka(); err != nil {
		log.Fatalf("Failed to initialize Kafka: %v", err)
	}

	handler := func(msg kafkago.Message) {
		log.Printf("Rebalancing Portfolio: %s", string(msg.Value))
		handlers.HandleRebalanceConsume(ctx, msg.Value)
	}
	kafka.ConsumeMessage(ctx, handler)

	http.HandleFunc("/portfolio", handlers.HandlePortfolio)
	http.HandleFunc("/rebalance", handlers.HandleRebalance)
	http.HandleFunc("/rebalance/list", handlers.HandleGetRebalance)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
