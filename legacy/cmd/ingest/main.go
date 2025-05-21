package main

import (
	"log"
	"os"
	"strings"

	"remote-patient-monitoring-system/internal/application/ingest"
	httpHandlers "remote-patient-monitoring-system/internal/infrastructure/http"
	"remote-patient-monitoring-system/internal/infrastructure/influxdb"
	"remote-patient-monitoring-system/internal/infrastructure/kafka"

	"github.com/gin-gonic/gin"
)

func main() {
	// environment config
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	topic := os.Getenv("OBS_TOPIC")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	ingestPort := os.Getenv("INGEST_PORT")
	if ingestPort == "" {
		ingestPort = "8081"
		log.Printf("INGEST_PORT not set, defaulting to %s", ingestPort)
	}

	// initialize influxdb repo
	obsRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("cannot initialize InfluxDB repo: %v", err)
	}

	// initializing kafka publisher
	pub := kafka.NewKafkaPublisher(brokers, topic)

	// create of ingest service and HTTP handler
	svc := ingest.NewIngestService(pub, obsRepo)
	ingestHandler := httpHandlers.NewIngestHandler(svc)

	// start HTTP server
	router := gin.Default()

	// register routes
	ingestHandler.RegisterRoutes(router)

	log.Printf("Ingest service listening on :%s", ingestPort)
	if err := router.Run(":" + ingestPort); err != nil {
		log.Fatal(err)
	}
}
