package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"remote-patient-monitoring-system/internal/application/query"
	"remote-patient-monitoring-system/internal/domain/model"
	httpHandlers "remote-patient-monitoring-system/internal/infrastructure/http"
	"remote-patient-monitoring-system/internal/infrastructure/influxdb"
	"remote-patient-monitoring-system/internal/infrastructure/kafka"
	"remote-patient-monitoring-system/internal/infrastructure/postgres"
	"remote-patient-monitoring-system/internal/infrastructure/ws"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// --- environment configuration ---
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	alertTopic := os.Getenv("ALERT_TOPIC")
	apiPort := os.Getenv("API_PORT")
	groupID := os.Getenv("GROUP_ID")

	// --- start repositories ---
	metricRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatal("error starting InfluxDB:", err)
	}

	alertRepo, err := postgres.NewPostgresRepo(conn)
	if err != nil {
		log.Fatal("error starting Postgres:", err)
	}

	// --- start services ---
	querySvc := query.NewQueryService(metricRepo, alertRepo)

	// --- start handlers ---
	queryHandler := httpHandlers.NewQueryHandler(querySvc)

	// --- start WebSocket ---
	wsHandler := ws.NewWSHandler()

	// --- subscribe to alerts topic ---
	go func() {
		consumer := kafka.NewKafkaConsumer(brokers, alertTopic, groupID)
		consumer.Consume(context.Background(), func(key, value []byte) {
			var alert model.Alert
			if err := json.Unmarshal(value, &alert); err != nil {
				log.Println("Mensaje de alerta inválido:", err)
				return
			}
			wsHandler.BroadcastAlert(&alert)
		})
	}()

	// --- router config ---
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// register handlers routes
	api := router.Group("/")
	queryHandler.RegisterRoutes(api)

	// websocket endpoint to receive alerts
	router.GET("/ws/alerts", gin.WrapF(wsHandler.Handler()))

	// --- start server ---
	log.Printf("API service listening on en :%s\n", apiPort)
	if err := router.Run(":" + apiPort); err != nil {
		log.Fatal(err)
	}
}
