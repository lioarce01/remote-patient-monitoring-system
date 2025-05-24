package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/repository"
)

// TelemetryInput represents unprocessed data
type TelemetryInput struct {
	PatientID string    `json:"patient_id"`
	Type      string    `json:"type"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type IngestService struct {
	Publisher       repository.Publisher
	ObservationRepo repository.ObservationRepository
	AlertRepo       repository.AlertRepository
	validator       *Validator
	normalizer      *Normalizer
}

func NewIngestService(pub repository.Publisher, obsRepo repository.ObservationRepository) *IngestService {
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
	// capture any internal panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in IngestService.Execute: %v", r)
		}
	}()

	// entry logs
	log.Printf("[Ingest] Execute called – input: %+v", input)

	// normalize data
	obs := svc.normalizer.FromTelemetry(input)
	log.Printf("[Ingest] Normalized obs: %+v", obs)
	if obs == nil {
		return errors.New("observation is nil after normalization")
	}

	// assign unique ID
	obs.ID = fmt.Sprintf("obs-%d", time.Now().UnixNano())
	log.Printf("[Ingest] Assigned ID: %s", obs.ID)

	// convert to flat record
	record, err := entities.ToObservationRecord(obs)
	log.Printf("[Ingest] ToObservationRecord returned: %+v, err: %v", record, err)
	if err != nil {
		return fmt.Errorf("conversion error: %w", err)
	}

	obsFHIR := entities.Observation{
		ID:                record.ID,
		ResourceType:      record.ResourceType,
		Status:            record.Status,
		Code:              entities.Code{Text: record.CodeText},
		Subject:           entities.Subject{Reference: record.PatientID},
		EffectiveDateTime: record.EffectiveDateTime.Format(time.RFC3339),
		ValueQuantity:     entities.ValueQuantity{Value: record.Value, Unit: record.Unit},
	}

	// serialize and publish FHIR observation
	payload, err := json.Marshal(obsFHIR)
	if err != nil {
		return fmt.Errorf("failed to marshal Observation: %w", err)
	}

	// publish on kafka
	if err := svc.Publisher.PublishFHIR(ctx, payload); err != nil {
		return fmt.Errorf("publish FHIR error: %w", err)
	}
	log.Println("[Ingest] Published FHIR Observation successfully")

	// save on influxdb
	log.Printf("[Ingest] Saving observation record to repository")
	if err := svc.ObservationRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("save error: %w", err)
	}
	log.Printf("[Ingest] Saved successfully")

	return nil
}
