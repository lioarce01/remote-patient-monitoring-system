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

	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/infrastructure/db"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/infrastructure/kafka"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/infrastructure/mlclient"
	"github.com/lioarce01/remote_patient_monitoring_system/processing-service/internal/application"
)

func main() {
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	obsTopic := os.Getenv("OBS_TOPIC")
	alertTopic := os.Getenv("ALERT_TOPIC")
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	groupID := os.Getenv("GROUP_ID")
	mlClient := mlclient.NewClient("http://ml-service:8000")

	// initialize kafka consumer
	consumer := kafka.NewKafkaConsumer(brokers, obsTopic, groupID)

	obsRepo, err := db.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("error initializing InfluxDB: %v", err)
	}

	alertRepo, err := db.NewPostgresRepo(conn)
	if err != nil {
		log.Fatalf("error initializing Postgres: %v", err)
	}

	// initialize publisher
	publisher := kafka.NewKafkaPublisher(brokers, alertTopic)

	// initialize processing service
	processingService := application.NewProcessService(publisher, alertRepo, obsRepo, mlClient)

	// context configuration to handler signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// initialize message consumption
	go func() {
		consumer.Consume(ctx, func(ket, msg []byte) {
			var obs entities.Observation
			if err := json.Unmarshal(msg, &obs); err != nil {
				log.Printf("invalid observation message: %v", err)
				return
			}

			record, err := entities.ToObservationRecord(&obs)
			if err != nil {
				log.Printf("error converting to ObservationRecord: %v", err)
				return
			}

			if err := processingService.HandleObservation(ctx, record); err != nil {
				log.Printf("error processing observation %s: %v", obs.ID, err)
			}
		})
	}()

	// signal finisher
	<-ctx.Done()
	log.Printf("signal finisher received, waiting to finalize processes...")
	time.Sleep(2 * time.Second)
	log.Println("processing service stopped")
}
