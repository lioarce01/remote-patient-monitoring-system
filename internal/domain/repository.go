package domain

import (
	"context"
	"remote-patient-monitoring-system/internal/domain/model"
)

type ObservationRepository interface {
	Save(ctx context.Context, record *model.ObservationRecord) error
	FetchObservations(ctx context.Context, patientID, from, to string) ([]model.Observation, error)
}

type AlertRepository interface {
	Save(ctx context.Context, alert *model.Alert) error
	FetchByPatient(ctx context.Context, patientID string) ([]model.Alert, error)
}

type Publisher interface {
	PublishObservation(ctx context.Context, obs *model.ObservationRecord) error
	PublishAlert(ctx context.Context, alert *model.Alert) error
}
