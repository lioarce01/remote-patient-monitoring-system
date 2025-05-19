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
	// --- Configuración ---
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	alertTopic := os.Getenv("ALERT_TOPIC")
	apiPort := os.Getenv("API_PORT")
	groupID := os.Getenv("GROUP_ID")

	// --- Inicializar repositorios ---
	metricRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatal("Error al inicializar InfluxDB:", err)
	}

	alertRepo, err := postgres.NewPostgresRepo(conn)
	if err != nil {
		log.Fatal("Error al inicializar Postgres:", err)
	}

	// --- Inicializar servicios ---
	querySvc := query.NewQueryService(metricRepo, alertRepo)

	// --- Inicializar handlers ---
	queryHandler := httpHandlers.NewQueryHandler(querySvc)

	// --- Inicializar WebSocket ---
	wsHandler := ws.NewWSHandler()

	// --- Suscribirse al tópico de alertas ---
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

	// --- Configurar router ---
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Registrar rutas de handlers
	api := router.Group("/")
	queryHandler.RegisterRoutes(api)

	// Endpoint WebSocket para recibir alertas en tiempo real
	router.GET("/ws/alerts", gin.WrapF(wsHandler.Handler()))

	// --- Iniciar servidor ---
	log.Printf("API service listening on en :%s\n", apiPort)
	if err := router.Run(":" + apiPort); err != nil {
		log.Fatal(err)
	}
}
