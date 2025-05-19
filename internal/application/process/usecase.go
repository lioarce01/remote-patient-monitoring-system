package process

import (
	"context"
	"fmt"
	"log"
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
		ZDetector:      rules.NewZScoreDetector(30, 3.0, 0.1),
		Thresholds:     &rules.Thresholds{HeartRateMax: 100, SpO2Min: 90},
	}
}

func (svc *ProcessService) HandleObservation(ctx context.Context, obs *model.ObservationRecord) error {
	// Chequear los umbrales y generar una alerta si es necesario
	if alertType, triggered := rules.CheckThresholds(obs, svc.Thresholds); triggered {
		log.Printf("Anomalía detectada: %s para paciente %s valor %f", alertType, obs.PatientID, obs.Value)

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

	// detector de anomalias
	if svc.ZDetector.Add(obs.Value) {
		alert := model.Alert{
			ID:            svc.generateID(),
			PatientID:     obs.PatientID,
			ObservationID: obs.ID,
			Type:          "Anomaly",
			Message:       fmt.Sprintf("Anomaly detected: value=%.2f at %s", obs.Value, obs.EffectiveDateTime),
			Timestamp:     time.Now(),
		}

		log.Printf("ZScore anomaly detected para paciente %s valor %f", obs.PatientID, obs.Value)

		if err := svc.publishAndSaveAlert(ctx, &alert); err != nil {
			return fmt.Errorf("failed to handle anomaly alert: %v", err)
		} else {
			log.Printf("Alerta anómala guardada y publicada con éxito para paciente %s", obs.PatientID)
		}
	}

	// Guardar métricas agregadas
	if err := svc.MetricsRepo.Save(ctx, obs); err != nil {
		return fmt.Errorf("failed to store metrics for patient %s: %v", obs.PatientID, err)
	}

	return nil
}

func (svc *ProcessService) publishAndSaveAlert(ctx context.Context, alert *model.Alert) error {
	log.Printf("Publicando alerta tipo %s para paciente %s", alert.Type, alert.PatientID)
	if err := svc.AlertPublisher.PublishAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to publish alert: %v", err)
	}
	log.Printf("Alerta publicada, guardando en Postgres...")
	if err := svc.AlertRepo.Save(ctx, alert); err != nil {
		return fmt.Errorf("failed to save alert: %v", err)
	}
	log.Printf("Alerta guardada en Postgres con id %s", alert.ID)
	return nil
}
