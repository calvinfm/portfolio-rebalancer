package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// InitKafka initializes kafka connection
func InitKafka() error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		return nil // skip if env not set
	}

	writer = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	// Retry logic to check Kafka availability
	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(context.Background(), kafka.Message{
			Value: []byte("ping"),
		})
		if err == nil {
			log.Println("Kafka is ready")
			return nil
		}
		log.Println("Waiting for Kafka to be ready...")
		time.Sleep(2 * time.Second)
	}

	return nil
}

func PublishMessage(ctx context.Context, payload []byte) error {
	if writer == nil {
		log.Println("Kafka writer is nil; skipping message publish")
		return fmt.Errorf("kafka writer not initialized")
	}

	msg := kafka.Message{
		Value: payload,
	}

	return writer.WriteMessages(ctx, msg)
}

func ConsumeMessage(ctx context.Context, handler func(kafka.Message)) error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		return nil
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		GroupID:  "portfolio-rebalancer-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	go func() {
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("Kafka consumer stopping...")
					return
				}
				log.Printf("Kafka read error: %v\n", err)
				time.Sleep(time.Second)
				continue
			}

			handler(msg)
		}
	}()

	log.Println("Kafka consumer started")
	return nil
}
