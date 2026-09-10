package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gotracker/internal/order"

	"github.com/segmentio/kafka-go"
)

var topic = "orders"

func SendOrderCreated(o order.Order) error {
	broker := os.Getenv("KAFKA_BROKER")

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{
			broker,
		},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	defer writer.Close()

	data, err := json.Marshal(o)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(fmt.Sprint(o.ID)),
		Value: data,
		Time:  time.Now(),
	}

	return writer.WriteMessages(
		context.Background(),
		message,
	)
}
