package ingest

import (
	"context"
	"errors"
	"remote-patient-monitoring-system/internal/domain"
	"time"
)

// TelemetryInput representa datos sin procesar.
type TelemetryInput struct {
	PatientID string    `json:"patient_id"`
	HeartRate float64   `json:"heart_rate"`
	Timestamp time.Time `json:"timestamp"`
}

// IngestService orquesta validación, normalización y persistencia.
type IngestService struct {
	Publisher       domain.Publisher
	ObservationRepo domain.ObservationRepository
	validator       *Validator
	normalizer      *Normalizer
}

func NewIngestService(pub domain.Publisher, obsRepo domain.ObservationRepository) *IngestService {
	return &IngestService{
		Publisher:       pub,
		ObservationRepo: obsRepo,
		validator:       NewValidator(),
		normalizer:      NewNormalizer(),
	}
}

func (svc *IngestService) Execute(ctx context.Context, input TelemetryInput) error {
	obs := svc.normalizer.FromTelemetry(input)
	if err := svc.validator.Validate(obs); err != nil {
		return errors.New("invalid observation: " + err.Error())
	}
	if err := svc.Publisher.PublishObservation(ctx, obs); err != nil {
		return err
	}
	return svc.ObservationRepo.Save(ctx, obs)
}
