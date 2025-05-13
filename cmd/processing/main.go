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
	// --- Configuración desde entorno ---
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	obsTopic := os.Getenv("OBS_TOPIC")
	alertTopic := os.Getenv("ALERT_TOPIC")
	conn := os.Getenv("POSTGRES_CONN")
	influxAddr := os.Getenv("INFLUX_ADDR")
	influxDB := os.Getenv("INFLUX_DB")
	influxUser := os.Getenv("INFLUX_USER")
	influxPass := os.Getenv("INFLUX_PASS")
	groupID := os.Getenv("GROUP_ID")

	// --- Inicializar consumidor de Kafka ---
	consumer := kafka.NewKafkaConsumer(brokers, obsTopic, groupID)

	// --- Inicializar repositorios ---
	metricRepo, err := influxdb.NewInfluxRepo(influxAddr, influxDB, influxUser, influxPass)
	if err != nil {
		log.Fatalf("Error al inicializar InfluxDB: %v", err)
	}

	alertRepo, err := postgres.NewPostgresRepo(conn)
	if err != nil {
		log.Fatalf("Error al inicializar Postgres: %v", err)
	}

	// --- Inicializar publisher ---
	publisher := kafka.NewKafkaPublisher(brokers, alertTopic)

	// --- Inicializar servicio de procesamiento ---
	svc := process.NewProcessService(publisher, alertRepo, metricRepo)

	// --- Configurar contexto para manejo de señales ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- Iniciar consumo de mensajes ---
	go func() {
		consumer.Consume(ctx, func(key, msg []byte) {
			var obs model.Observation
			if err := json.Unmarshal(msg, &obs); err != nil {
				log.Printf("Mensaje de observación inválido: %v", err)
				return
			}

			record, err := model.ToObservationRecord(&obs)
			if err != nil {
				log.Printf("Error al convertir a ObservationRecord: %v", err)
				return
			}

			if err := svc.HandleObservation(ctx, record); err != nil {
				log.Printf("Error al procesar observación %s: %v", obs.ID, err)
			}
		})
	}()

	// --- Esperar señal de terminación ---
	<-ctx.Done()
	log.Println("Señal de terminación recibida, esperando finalización de procesos...")
	time.Sleep(2 * time.Second)
	log.Println("Servicio de procesamiento detenido")
}
