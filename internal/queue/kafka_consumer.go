package queue

import (
	"context"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

func StartConsumer() {
	go func() {
		broker := os.Getenv("KAFKA_BROKER")

		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{
				broker,
			},
			Topic:   "orders",
			GroupID: "gotracker-group",
		})

		defer reader.Close()

		fmt.Println("[Kafka] Consumer запущен")

		for {
			message, err := reader.ReadMessage(
				context.Background(),
			)

			if err != nil {
				fmt.Println(
					"[Kafka] Ошибка:",
					err,
				)
				continue
			}

			fmt.Printf(
				"[Kafka] Новый заказ: %s\n",
				string(message.Value),
			)
		}
	}()
}
