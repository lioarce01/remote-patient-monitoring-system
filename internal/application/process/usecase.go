package process

import (
	"context"
	"fmt"
	"remote-patient-monitoring-system/internal/domain"
	"remote-patient-monitoring-system/internal/domain/model"
	"remote-patient-monitoring-system/internal/domain/rules"
	"time"
)

type ProcessService struct {
	AlertPublisher domain.Publisher
	AlertRepo      domain.AlertRepository       // writes to Postgres
	MetricsRepo    domain.ObservationRepository // aggregated metrics (Postgres)
	ZDetector      *rules.ZScoreDetector
	Thresholds     *rules.Thresholds
}

// generateID generates a unique identifier for alerts.
func (svc *ProcessService) generateID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

func NewProcessService(publisher domain.Publisher, alertRepo domain.AlertRepository, metricsRepo domain.ObservationRepository) *ProcessService {
	return &ProcessService{
		AlertPublisher: publisher,
		AlertRepo:      alertRepo,
		MetricsRepo:    metricsRepo,
		ZDetector:      rules.NewZScoreDetector(50, 3.0), // 50-point window, 3σ
		Thresholds:     &rules.Thresholds{HeartRateMax: 100, SpO2Min: 90},
	}
}

func (svc *ProcessService) HandleObservation(ctx context.Context, obs *model.Observation) error {
	if alertType, triggered := rules.CheckThresholds(obs, svc.Thresholds); triggered {
		alert := model.Alert{
			ID:        svc.generateID(),
			PatientID: obs.PatientID,
			Type:      alertType,
			Message:   fmt.Sprintf("%s: value=%.2f at %s", alertType, obs.Value, obs.Timestamp),
			Timestamp: time.Now(),
		}
		// Publish alert event to Kafka and save to DB.
		if err := svc.AlertPublisher.PublishAlert(ctx, &alert); err != nil {
			return err
		}
		if err := svc.AlertRepo.Save(ctx, &alert); err != nil {
			return err
		}
	}
	// Example anomaly detection:
	if svc.ZDetector.Add(obs.Value) {
		alert := model.Alert{ /* as above, type "Anomaly" */ }
		svc.AlertPublisher.PublishAlert(ctx, &alert)
		svc.AlertRepo.Save(ctx, &alert)
	}
	// Store aggregated metric (could be in Postgres as well)
	if err := svc.MetricsRepo.Save(ctx, obs); err != nil {
		return err
	}
	return nil
}
