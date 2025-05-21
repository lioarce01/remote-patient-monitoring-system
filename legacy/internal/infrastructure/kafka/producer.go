package kafka

import (
	"context"
	"encoding/json"
	"log"
	"remote-patient-monitoring-system/internal/domain/model"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	w     *kafka.Writer
	topic string
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	if len(brokers) == 0 || topic == "" {
		log.Fatalf("kafka: brokers and topic are required")
	}
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	return &KafkaPublisher{w: w, topic: topic}
}

func (p *KafkaPublisher) PublishObservation(ctx context.Context, obs *model.ObservationRecord) error {
	msg, _ := json.Marshal(obs)
	return p.w.WriteMessages(ctx, kafka.Message{Key: []byte(obs.PatientID), Value: msg})
}

func (p *KafkaPublisher) PublishAlert(ctx context.Context, alert *model.Alert) error {
	msg, _ := json.Marshal(alert)
	return p.w.WriteMessages(ctx, kafka.Message{Key: []byte(alert.PatientID), Value: msg})
}

func (p *KafkaPublisher) PublishFHIR(ctx context.Context, payload []byte) error {
	return p.w.WriteMessages(ctx, kafka.Message{
		Key:   nil,
		Value: payload,
	})
}
