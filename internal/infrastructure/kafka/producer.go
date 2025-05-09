package kafka

import (
	"context"
	"encoding/json"
	"remote-patient-monitoring-system/internal/domain/model"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct{ w *kafka.Writer }

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{w: kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers, Topic: topic, Balancer: &kafka.Hash{},
	})}
}
func (p *KafkaPublisher) PublishObservation(ctx context.Context, obs *model.Observation) error {
	msg, _ := json.Marshal(obs)
	return p.w.WriteMessages(ctx, kafka.Message{Key: []byte(obs.PatientID), Value: msg})
}

func (p *KafkaPublisher) PublishAlert(ctx context.Context, alert *model.Alert) error {
	msg, _ := json.Marshal(alert)
	return p.w.WriteMessages(ctx, kafka.Message{Key: []byte(alert.PatientID), Value: msg})
}
