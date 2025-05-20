# Remote Patient Monitoring System

A real-time Remote Patient Monitoring (RPM) system designed to collect, process, and visualize patient telemetry data. Built with Go, Kafka, InfluxDB, PostgreSQL, Prometheus, and WebSockets, and deployed via Docker Compose.

---

## Table of Contents

* [Introduction](#introduction)
* [Architecture](#architecture)
* [Tech Stack](#tech-stack)
* [Getting Started](#getting-started)

  * [Clone the Repository](#clone-the-repository)
  * [Configure Environment Variables](#configure-environment-variables)
  * [Run with Docker Compose](#run-with-docker-compose)
* [Project Structure](#project-structure)
* [Services](#services)

  * [Ingest Service](#ingest-service)
  * [Processing Service](#processing-service)
  * [API Service](#api-service)
* [Environment Configuration](#environment-configuration)

  * [Shared Variables](#shared-variables)
  * [Service-Specific Variables](#service-specific-variables)
* [Usage](#usage)

  * [REST Endpoints](#rest-endpoints)
  * [WebSocket Notifications](#websocket-notifications)
* [Monitoring & Metrics](#monitoring--metrics)

---

## Introduction

This project implements a scalable RPM system that captures patient vital signs and telemetry, processes the data in real-time, stores metrics and alerts in specialized databases, and exposes APIs and WebSockets for data retrieval and notifications.

## Architecture

1. **Ingest Service**: Receives raw patient observations via HTTP and publishes messages to a Kafka topic.
2. **Processing Service**: Consumes observations from Kafka, applies business rules, writes metrics to InfluxDB, stores generated alerts in PostgreSQL, and publishes alerts to a Kafka topic.
3. **API Service**: Exposes REST endpoints to query historical observations and alerts, and a WebSocket endpoint for real-time alert streaming.

## Tech Stack

* **Language**: Go (Golang)
* **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
* **WebSockets**: [Gorilla WebSocket](https://github.com/gorilla/websocket)
* **Databases**:

  * InfluxDB (time-series data)
  * PostgreSQL (relational data)
* **Message Broker**: Apache Kafka
* **ORM**: [GORM](https://gorm.io/)
* **Monitoring**: Prometheus
* **Deployment**: Docker & Docker Compose

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/lioarce01/remote-patient-monitoring-system.git
cd remote-patient-monitoring-system
```

### Configure Environment Variables

Copy the `.env.example` file and update values:

```bash
cp .env.example .env
# Edit .env with your configuration (ports, DB credentials, Kafka brokers)
```

### Run with Docker Compose

```bash
docker-compose up --build
```

All services will start:

* **InfluxDB**: [http://localhost:8086](http://localhost:8086)
* **PostgreSQL**: port as defined in `.env`
* **Kafka** & **Zookeeper**: localhost:9092 & 2181
* **API Service**: [http://localhost:\${API\_PORT}](http://localhost:${API_PORT})
* **Prometheus**: [http://localhost:9090](http://localhost:9090)

## Project Structure

```
remote-patient-monitoring-system/
├── cmd/
│   ├── api/           # API service entrypoint
│   ├── ingest/        # Ingest service entrypoint
│   └── processing/    # Processing service entrypoint
├── internal/
│   ├── application/   # Business logic and use cases
│   ├── domain/        # Core domain models
   │   └── infrastructure/ # Kafka, DB clients, repositories
├── scripts/              # Diagrams and additional documentation
├── docker-compose.yml
├── Dockerfile         # Multi-stage build for Go services
├── go.mod
```

## Services

### Ingest Service

* Listens on `INGEST_PORT`

* Accepts HTTP POST requests with JSON payload:

  ```json
  {
    "patient_id":"Patient123",
    "type":"heart_rate",
    "value":100,
    "unit":"bpm",
    "timestamp":"2025-05-17T18:03:00Z"
  }
  ```
  
* Publishes to Kafka topic defined by `OBS_TOPIC`.

### Processing Service

* Consumes messages from `OBS_TOPIC`
* Applies thresholds (configurable via code or env)
* Writes time-series points to InfluxDB
* If metrics exceed thresholds, generates an alert record in PostgreSQL and publishes to `ALERT_TOPIC`.

### API Service

* REST API for reading data:

  * `GET /observations?patient_id={id}&from={ts}&to={ts}`
  * `GET /alerts?patient_id={id}`
* WebSocket endpoint:

  * `ws://localhost:${API_PORT}/ws/alerts` for real-time alert streaming

## Environment Configuration

### Shared Variables

```dotenv
INFLUX_ADDR=http://influxdb:8086
INFLUX_DB=rpm_metrics
INFLUX_USER=admin
INFLUX_PASS=secret
POSTGRES_CONN=postgres://user:pass@postgres:5432/rpm_alerts?sslmode=disable
KAFKA_BROKERS=broker:9092
OBS_TOPIC=observations
ALERT_TOPIC=alerts
```

### Service-Specific Variables

```dotenv
# API Service
API_PORT=8080

# Ingest Service
INGEST_PORT=8081

# Processing Service
# (can use same ports for health checks, metrics)
```

## Usage

### REST Endpoints

* Fetch observations:

  ```bash
  curl "http://localhost:8080/observations?patient_id=123&from=2025-05-01T00:00:00Z&to=2025-05-16T23:59:59Z"
  ```

* Fetch alerts:

  ```bash
  curl "http://localhost:8080/alerts?patient_id=123"
  ```

### WebSocket Notifications

Connect to the WebSocket endpoint for live alerts:

```bash
wscat -c ws://localhost:8080/ws/alerts
```

Messages will be sent in JSON format:

```json
{
  "ID": "alert-1747693382984800835",
  "PatientID": "Patient123",
  "ObservationID": "obs-1747693381973770098",
  "Message": "Anomaly detected: value=127.00 at 2025-05-19 22:30:00 +0000 UTC",
  "Type": "Anomaly",
  "Timestamp": "2025-05-19T22:23:02.984849133Z",
  "Acknowledged": false
}
```

## Monitoring & Metrics

* Prometheus scrapes metrics from each service on `/metrics` (default port)
