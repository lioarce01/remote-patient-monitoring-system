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

func (svc *ProcessService) generateID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

func NewProcessService(publisher domain.Publisher, alertRepo domain.AlertRepository, metricsRepo domain.ObservationRepository) *ProcessService {
	return &ProcessService{
		AlertPublisher: publisher,
		AlertRepo:      alertRepo,
		MetricsRepo:    metricsRepo,
		ZDetector:      rules.NewZScoreDetector(50, 3.0),
		Thresholds:     &rules.Thresholds{HeartRateMax: 100, SpO2Min: 90},
	}
}

func (svc *ProcessService) HandleObservation(ctx context.Context, obs *model.ObservationRecord) error {
	// Chequear los umbrales y generar una alerta si es necesario
	if alertType, triggered := rules.CheckThresholds(obs, svc.Thresholds); triggered {
		alert := model.Alert{
			ID:            svc.generateID(),
			PatientID:     obs.PatientID,
			ObservationID: obs.ID, // Relacionamos la alerta con la observación
			Type:          alertType,
			Message:       fmt.Sprintf("%s: value=%.2f at %s", alertType, obs.Value, obs.EffectiveDateTime),
			Timestamp:     time.Now(),
		}

		// Publicar y guardar alerta
		if err := svc.publishAndSaveAlert(ctx, &alert); err != nil {
			return fmt.Errorf("failed to handle alert for patient %s: %v", alert.PatientID, err)
		}
	}

	// Detección de anomalías
	if svc.ZDetector.Add(obs.Value) {
		alert := model.Alert{
			ID:            svc.generateID(),
			PatientID:     obs.PatientID,
			ObservationID: obs.ID, // Relacionamos la alerta con la observación
			Type:          "Anomaly",
			Message:       fmt.Sprintf("Anomaly detected: value=%.2f at %s", obs.Value, obs.EffectiveDateTime),
			Timestamp:     time.Now(),
		}
		// Publicar y guardar alerta de anomalía
		if err := svc.publishAndSaveAlert(ctx, &alert); err != nil {
			return fmt.Errorf("failed to handle anomaly alert: %v", err)
		}
	}

	// Guardar métricas agregadas
	if err := svc.MetricsRepo.Save(ctx, obs); err != nil {
		return fmt.Errorf("failed to store metrics for patient %s: %v", obs.PatientID, err)
	}

	return nil
}

func (svc *ProcessService) publishAndSaveAlert(ctx context.Context, alert *model.Alert) error {
	// Publicar la alerta a Kafka
	if err := svc.AlertPublisher.PublishAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to publish alert: %v", err)
	}
	// Guardar la alerta en la base de datos
	if err := svc.AlertRepo.Save(ctx, alert); err != nil {
		return fmt.Errorf("failed to save alert: %v", err)
	}
	return nil
}
