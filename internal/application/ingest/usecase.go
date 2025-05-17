package ingest

import (
	"context"
	"encoding/json"
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
	AlertRepo       domain.AlertRepository
	validator       *Validator
	normalizer      *Normalizer
}

func NewIngestService(pub domain.Publisher, obsRepo domain.ObservationRepository) *IngestService {
	if pub == nil || obsRepo == nil {
		log.Fatal("Publisher, ObservationRepo or AlertRepo is nil")
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

	obsFHIR := model.Observation{
		ID:                record.ID,
		ResourceType:      record.ResourceType,
		Status:            record.Status,
		Code:              model.Code{Text: record.CodeText},
		Subject:           model.Subject{Reference: record.PatientID},
		EffectiveDateTime: record.EffectiveDateTime.Format(time.RFC3339),
		ValueQuantity:     model.ValueQuantity{Value: record.Value, Unit: record.Unit},
	}

	// Serializa y publica el FHIR Observation
	payload, err := json.Marshal(obsFHIR)
	if err != nil {
		return fmt.Errorf("failed to marshal Observation: %w", err)
	}

	// 6) Publicar en Kafka
	if err := svc.Publisher.PublishFHIR(ctx, payload); err != nil {
		return fmt.Errorf("publish FHIR error: %w", err)
	}
	log.Println("[Ingest] Published FHIR Observation successfully")

	// 7) Guardar en InfluxDB
	log.Printf("[Ingest] Saving observation record to repository")
	if err := svc.ObservationRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("save error: %w", err)
	}
	log.Printf("[Ingest] Saved successfully")

	return nil
}
