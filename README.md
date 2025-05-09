# Remote Patient Monitoring System

This project is a real-time Remote Patient Monitoring (RPM) system designed to collect, process, and visualize patient telemetry data. The architecture is composed of three main services:

- **Ingest Service**: Receives patient observation data and publishes it to Kafka.
- **Processing Service**: Consumes data from Kafka, processes it, and generates alerts if necessary.
- **API Service**: Exposes a REST API to query observations and alerts, and provides a WebSocket for real-time alert notifications.

## 🚀 Tech Stack

- **Programming Language**: Go (Golang)
- **Databases**:
  - InfluxDB: Stores patient observation metrics
  - PostgreSQL: Stores alerts
- **Message Broker**: Apache Kafka
- **Web Framework**: Gin
- **WebSocket Library**: Gorilla WebSocket
- **Monitoring**: Prometheus

- ## 📁 Project Structure

```
remote-patient-monitoring-system/
├── cmd/
│   ├── api/           # API service
│   ├── ingest/        # Ingest service
│   └── processing/    # Processing service
├── internal/
│   ├── application/   # Business logic
│   ├── domain/        # Domain models
│   └── infrastructure/# Infra implementations (Kafka, DBs, etc.)
├── go.mod
└── README.md
```

## ⚙️ Environment Configuration

Set the following environment variables for each service:

### Shared Variables

- `INFLUX_ADDR`: InfluxDB address (e.g., `http://localhost:8086`)
- `INFLUX_DB`: InfluxDB database name
- `INFLUX_USER`: InfluxDB username
- `INFLUX_PASS`: InfluxDB password
- `POSTGRES_CONN`: PostgreSQL connection string
- `KAFKA_BROKERS`: Comma-separated list of Kafka brokers (e.g., `localhost:9092`)
- `OBS_TOPIC`: Kafka topic for observations
- `ALERT_TOPIC`: Kafka topic for alerts

### Service-Specific Variables

- **API Service**:
  - `API_PORT`: Port to run the API service (e.g., `8080`)
- **Ingest Service**:
  - `INGEST_PORT`: Port to run the ingest service (e.g., `8081`)
