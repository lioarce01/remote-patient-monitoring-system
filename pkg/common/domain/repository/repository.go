package repository

import (
	"context"

	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
)

type ObservationRepository interface {
	Save(ctx context.Context, record *entities.ObservationRecord) error
	FetchObservations(ctx context.Context, patientID, from, to string) ([]entities.Observation, error)
}

type AlertRepository interface {
	Save(ctx context.Context, alert *entities.Alert) error
	FetchByPatient(ctx context.Context, patientID string) ([]entities.Alert, error)
}

type Publisher interface {
	PublishObservation(ctx context.Context, obs *entities.ObservationRecord) error
	PublishAlert(ctx context.Context, alert *entities.Alert) error
	PublishFHIR(ctx context.Context, payload []byte) error
}
