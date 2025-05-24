package query

import (
	"context"
	"remote-patient-monitoring-system/internal/domain"
	"remote-patient-monitoring-system/internal/domain/model"
)

type QueryService struct {
	MetricsRepo domain.ObservationRepository
	AlertRepo   domain.AlertRepository
}

func NewQueryService(mRepo domain.ObservationRepository, aRepo domain.AlertRepository) *QueryService {
	return &QueryService{MetricsRepo: mRepo, AlertRepo: aRepo}
}

func (s *QueryService) GetPatientObservations(ctx context.Context, patientID, from, to string) ([]model.Observation, error) {
	return s.MetricsRepo.FetchObservations(ctx, patientID, from, to)
}

func (s *QueryService) GetPatientAlerts(ctx context.Context, patientID string) ([]model.Alert, error) {
	return s.AlertRepo.FetchByPatient(ctx, patientID)
}
