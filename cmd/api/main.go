package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"remote-patient-monitoring-system/internal/application/query"
	"remote-patient-monitoring-system/internal/domain/model"
	"remote-patient-monitoring-system/internal/infrastructure/influxdb"
	"remote-patient-monitoring-system/internal/infrastructure/kafka"
	"remote-patient-monitoring-system/internal/infrastructure/postgres"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	alertTopic := os.Getenv("ALERTS_TOPIC")
	apiPort := os.Getenv("API_PORT")

	// --- Initialize repos ---
	// Initialize the InfluxDB repository
	metricRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatal("InfluxDB repo error:", err)
	}

	// Initialize the PostgreSQL repository
	alertRepo, err := postgres.NewPostgresRepo(conn)
	if err != nil {
		log.Fatal("Postgres repo error:", err)
	}

	// Initialize the Kafka producer
	svc := query.NewQueryService(metricRepo, alertRepo)

	// --- WebSocket hub ---
	var (
		upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		clients   = make(map[*websocket.Conn]bool)
		clientsMu sync.Mutex
	)

	// Broadcasts an alert to all connected clients.
	broadcastAlert := func(alert *model.Alert) {
		clientsMu.Lock()
		defer clientsMu.Unlock()
		for conn := range clients {
			if err := conn.WriteJSON(alert); err != nil {
				log.Println("WS: write error:", err)
				clientsMu.Lock()
				delete(clients, conn)
				clientsMu.Unlock()
				conn.Close()
			}
		}
	}

	// --- sub alerts topic ---
	go func() {
		consumer := kafka.NewKafkaConsumer(brokers, alertTopic, "api-alerts-group")
		consumer.Consume(context.Background(), func(key, value []byte) {
			var alert model.Alert
			if err := json.Unmarshal(value, &alert); err != nil {
				log.Println("WS: invalid alert msg:", err)
				return
			}
			broadcastAlert(&alert)
		})
	}()

	// --- Router Gin ---
	router := gin.Default()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/")
	api.GET("/patients/:id/observations", func(c *gin.Context) {
		id := c.Param("id")
		from := c.Query("from")
		to := c.Query("to")
		data, err := svc.GetPatientObservations(c.Request.Context(), id, from, to)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	})
	api.GET("/patients/:id/alerts", func(c *gin.Context) {
		id := c.Param("id")
		data, err := svc.GetPatientAlerts(c.Request.Context(), id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	})

	// Endpoint WebSocket to receive alerts in real time
	router.GET("/ws/alerts", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		clientsMu.Lock()
		clients[ws] = true
		clientsMu.Unlock()
		// Optional: manage client disconnection
		// This will close the connection when the client disconnects
		defer func() {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			ws.Close()
		}()
		for {
			if _, _, err := ws.NextReader(); err != nil {
				break
			}
		}
	})

	// --- Iniciar servidor ---
	log.Printf("API service listening on :%s\n", apiPort)
	if err := router.Run(":" + apiPort); err != nil {
		log.Fatal(err)
	}
}
