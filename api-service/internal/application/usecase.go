package application

import (
	"context"

	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/domain/entities"
	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/domain/repository"
)

type QueryService struct {
	MetricsRepo repository.ObservationRepository
	AlertRepo   repository.AlertRepository
}

func NewQueryService(mRepo repository.ObservationRepository, aRepo repository.AlertRepository) *QueryService {
	return &QueryService{MetricsRepo: mRepo, AlertRepo: aRepo}
}

func (s *QueryService) GetPatientObservations(ctx context.Context, patientID, from, to string) ([]entities.Observation, error) {
	return s.MetricsRepo.FetchObservations(ctx, patientID, from, to)
}

func (s *QueryService) GetPatientAlerts(ctx context.Context, patientID string) ([]entities.Alert, error) {
	return s.AlertRepo.FetchByPatient(ctx, patientID)
}
