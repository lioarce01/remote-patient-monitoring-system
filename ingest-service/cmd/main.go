package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	httpHandler "github.com/lioarce01/remote-patient-monitoring-system/ingest-service/internal/infrastructure/http"

	"github.com/gin-gonic/gin"
	"github.com/lioarce01/remote-patient-monitoring-system/ingest-service/internal/application"
	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/infrastructure/db"
	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/infrastructure/kafka"
)

func main() {
	// environment config
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	obsTopic := os.Getenv("OBS_TOPIC")
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
	obsRepo, err := db.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("cannot initialize InfluxDB repo: %v", err)
	}

	// initialize kafka publisher
	pub := kafka.NewKafkaPublisher(brokers, obsTopic)

	// initialize ingest service & http handler
	ingestService := application.NewIngestService(pub, obsRepo)
	ingestHandler := httpHandler.NewIngestHandler(ingestService)

	router := gin.Default()

	// register routes
	ingestHandler.RegisterRoutes(router)

	// healthcheck
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	log.Printf("Ingest service listening on: %s", ingestPort)
	if err := router.Run(":" + ingestPort); err != nil {
		log.Fatal(err)
	}
}
