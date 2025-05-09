package main

import (
	"log"
	"os"
	"strings"

	"remote-patient-monitoring-system/internal/application/ingest"
	"remote-patient-monitoring-system/internal/infrastructure/influxdb"
	"remote-patient-monitoring-system/internal/infrastructure/kafka"

	"github.com/gin-gonic/gin"
)

func main() {
	// Config variables
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	topic := os.Getenv("OBS_TOPIC")

	pub := kafka.NewKafkaPublisher(brokers, topic)
	obsRepo, err := influxdb.NewInfluxRepo(
		os.Getenv("INFLUX_ADDR"),
		os.Getenv("INFLUX_DB"),
		os.Getenv("INFLUX_USER"),
		os.Getenv("INFLUX_PASS"),
	)
	if err != nil {
		log.Fatal(err)
	}

	svc := ingest.NewIngestService(pub, obsRepo)
	router := gin.Default()
	router.POST("/observations", func(c *gin.Context) {
		var in ingest.TelemetryInput
		if err := c.BindJSON(&in); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := svc.Execute(c.Request.Context(), in); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Status(202)
	})
	router.Run(":" + os.Getenv("INGEST_PORT"))
}
