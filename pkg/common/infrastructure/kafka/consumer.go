package kafka

import (
	"context"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaConsumer struct{ r *kafka.Reader }

func NewKafkaConsumer(brokers []string, topic, group string) *KafkaConsumer {
	if topic == "" {
		log.Fatal("Kafka topic must be provided but is empty")
	}
	if group == "" {
		log.Fatal("Kafka group ID must be provided but is empty")
	}
	if len(brokers) == 0 {
		log.Fatal("Kafka brokers must be provided")
	}

	return &KafkaConsumer{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: group,
		}),
	}
}

func (c *KafkaConsumer) Consume(ctx context.Context, handler func(key, val []byte)) {
	for {
		msg, err := c.r.ReadMessage(ctx)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		handler(msg.Key, msg.Value)
	}
}
