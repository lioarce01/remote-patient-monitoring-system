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
	// --- environment config ---
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	obsTopic := os.Getenv("OBS_TOPIC")
	alertTopic := os.Getenv("ALERT_TOPIC")
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	groupID := os.Getenv("GROUP_ID")

	// --- initialize kafka consumer ---
	consumer := kafka.NewKafkaConsumer(brokers, obsTopic, groupID)

	// --- initialize repositories ---
	metricRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("error initializing influxdb: %v", err)
	}

	alertRepo, err := postgres.NewPostgresRepo(conn)
	if err != nil {
		log.Fatalf("error initializing postgres: %v", err)
	}

	// --- initialize publisher ---
	publisher := kafka.NewKafkaPublisher(brokers, alertTopic)

	// --- initialize process service ---
	svc := process.NewProcessService(publisher, alertRepo, metricRepo)

	// --- context configuration to handle signals ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- initialize message consumption ---
	go func() {
		consumer.Consume(ctx, func(key, msg []byte) {
			var obs model.Observation
			if err := json.Unmarshal(msg, &obs); err != nil {
				log.Printf("observation message invalid: %v", err)
				return
			}

			record, err := model.ToObservationRecord(&obs)
			if err != nil {
				log.Printf("error converting to observationRecord: %v", err)
				return
			}

			if err := svc.HandleObservation(ctx, record); err != nil {
				log.Printf("error processing observation %s: %v", obs.ID, err)
			}
		})
	}()

	// --- finisher signal ---
	<-ctx.Done()
	log.Println("finisher signal received, waiting finalizing processes...")
	time.Sleep(2 * time.Second)
	log.Println("processing service stopped")
}
