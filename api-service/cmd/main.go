package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/infrastructure/db"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/infrastructure/kafka"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/infrastructure/ws"
	"github.com/lioarce01/rempote_patient_monitoring_system/api-service/internal/application"
	httpHandler "github.com/lioarce01/rempote_patient_monitoring_system/api-service/internal/infrastructure/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	alertTopic := os.Getenv("ALERT_TOPIC")
	apiPort := os.Getenv("API_PORT")
	groupID := os.Getenv("GROUP_ID")

	// initialize repositories
	obsRepo, err := db.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("cannot initialize InfluxDB repo: %v", err)
	}

	alertRepo, err := db.NewPostgresRepo(conn)
	if err != nil {
		log.Fatalf("cannot initialize Postgres repo: %v", err)
	}

	// initialize services
	apiService := application.NewQueryService(obsRepo, alertRepo)

	// initialize handlers
	queryHandler := httpHandler.NewQueryHandler(apiService)

	// start websocket
	wsHandler := ws.NewWSHandler()

	go func() {
		consumer := kafka.NewKafkaConsumer(brokers, alertTopic, groupID)
		consumer.Consume(context.Background(), func(key, value []byte) {
			var alert entities.Alert
			if err := json.Unmarshal(value, &alert); err != nil {
				log.Println("Invalid alert message:", err)
				return
			}
			wsHandler.BroadcastAlert(&alert)
		})
	}()

	router := gin.Default()

	// prometheus route
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/")
	queryHandler.RegisterRoutes(api)

	// websocket endpoint
	router.GET("/ws/alerts", gin.WrapF(wsHandler.Handler()))

	// healthcheck
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// initialize server
	log.Printf("API service listening on :%s\n", apiPort)
	if err := router.Run(":" + apiPort); err != nil {
		log.Fatal(err)
	}
}
