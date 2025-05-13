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
	// Leer config
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

	// InfluxDB repo
	obsRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("cannot initialize InfluxDB repo: %v", err)
	}

	// Después de NewKafkaPublisher
	pub := kafka.NewKafkaPublisher(brokers, topic)

	// Crear servicio de ingestión y handler HTTP
	svc := ingest.NewIngestService(pub, obsRepo)
	ingestHandler := httpHandlers.NewIngestHandler(svc)

	// arrancar HTTP
	router := gin.Default()

	//registrar rutas y handlers
	ingestHandler.RegisterRoutes(router)

	log.Printf("Ingest service listening on :%s", ingestPort)
	if err := router.Run(":" + ingestPort); err != nil {
		log.Fatal(err)
	}
}
