package ingest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"remote-patient-monitoring-system/internal/domain"
	"remote-patient-monitoring-system/internal/domain/model"
	"time"
)

// TelemetryInput representa datos sin procesar.
type TelemetryInput struct {
	PatientID string    `json:"patient_id"`
	Type      string    `json:"type"`  // ej: "heart_rate"
	Value     float64   `json:"value"` // ej: 78.0
	Unit      string    `json:"unit"`  // ej: "bpm"
	Timestamp time.Time `json:"timestamp"`
}

type IngestService struct {
	Publisher       domain.Publisher
	ObservationRepo domain.ObservationRepository
	validator       *Validator
	normalizer      *Normalizer
}

func NewIngestService(pub domain.Publisher, obsRepo domain.ObservationRepository) *IngestService {
	if pub == nil || obsRepo == nil {
		log.Fatal("Publisher or ObservationRepo is nil")
	}

	return &IngestService{
		Publisher:       pub,
		ObservationRepo: obsRepo,
		validator:       NewValidator(),
		normalizer:      NewNormalizer(),
	}
}

func (svc *IngestService) Execute(ctx context.Context, input TelemetryInput) (err error) {
	// 1) Capturar cualquier panic interno
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in IngestService.Execute: %v", r)
		}
	}()

	// 2) Log de entrada
	log.Printf("[Ingest] Execute called – input: %+v", input)

	// 3) Normalizar
	obs := svc.normalizer.FromTelemetry(input)
	log.Printf("[Ingest] Normalized obs: %+v", obs) // Verificar el valor aquí
	if obs == nil {
		return errors.New("observation is nil after normalization")
	}

	// 4) Asignar un ID único
	obs.ID = fmt.Sprintf("obs-%d", time.Now().UnixNano())
	log.Printf("[Ingest] Assigned ID: %s", obs.ID)

	// 5) Convertir a record plano
	record, err := model.ToObservationRecord(obs)
	log.Printf("[Ingest] ToObservationRecord returned: %+v, err: %v", record, err)
	if err != nil {
		return fmt.Errorf("conversion error: %w", err)
	}

	// 6) Publicar en Kafka
	log.Printf("[Ingest] Publishing observation record for patient %s", record.PatientID)
	if err := svc.Publisher.PublishObservation(ctx, record); err != nil {
		return fmt.Errorf("publish error: %w", err)
	}
	log.Printf("[Ingest] Published successfully")

	// 7) Guardar en InfluxDB
	log.Printf("[Ingest] Saving observation record to repository")
	if err := svc.ObservationRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("save error: %w", err)
	}
	log.Printf("[Ingest] Saved successfully")

	return nil
}
