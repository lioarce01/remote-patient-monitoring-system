package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"remote-patient-monitoring-system/internal/application/process"
	"remote-patient-monitoring-system/internal/domain/model"
	"remote-patient-monitoring-system/internal/infrastructure/influxdb"
	"remote-patient-monitoring-system/internal/infrastructure/kafka"
	"remote-patient-monitoring-system/internal/infrastructure/postgres"
)

func main() {
	// --- Configuración desde entorno ---
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	obsTopic := os.Getenv("OBS_TOPIC")
	alertTopic := os.Getenv("ALERT_TOPIC")
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")

	// ---  KAFKA CONSUMER ---
	// Create a Kafka consumer to consume messages from the specified topic
	consumer := kafka.NewKafkaConsumer(brokers, obsTopic, "processor-group")

	metricRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("InfluxDB repo error: %v", err)
	}

	alertRepo, err := postgres.NewPostgresRepo(conn)
	if err != nil {
		log.Fatalf("Postgres repo error: %v", err)
	}

	publisher := kafka.NewKafkaPublisher(brokers, alertTopic)

	// --- processing service ---
	svc := process.NewProcessService(publisher, alertRepo, metricRepo)

	// --- cancel context on shutdown ---
	// Create a context that will be cancelled on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- start consuming messages ---
	// Start the Kafka consumer in a separate goroutine
	go func() {
		consumer.Consume(ctx, func(key, msg []byte) {
			var obs model.Observation
			if err := json.Unmarshal(msg, &obs); err != nil {
				log.Printf("Invalid observation message: %v", err)
				return
			}
			if err := svc.HandleObservation(ctx, &obs); err != nil {
				log.Printf("Error processing observation %s: %v", obs.ID, err)
			}
		})
	}()

	// wait for shutdown signal
	<-ctx.Done()
	// shutdown signal received
	log.Println("Shutdown signal received, waiting for in-flight messages...")
	time.Sleep(2 * time.Second)
	log.Println("Processing service stopped")
}
