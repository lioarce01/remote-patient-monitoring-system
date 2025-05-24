package application

import (
	"context"
	"fmt"
	"log"

	"strings"
	"time"

	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/domain/entities"
	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/domain/repository"
	"github.com/lioarce01/remote-patient-monitoring-system/pkg/common/infrastructure/mlclient"
	"github.com/lioarce01/remote-patient-monitoring-system/processing-service/internal/domain/rules"
)

type ProcessService struct {
	AlertPublisher repository.Publisher
	AlertRepo      repository.AlertRepository       // writes to Postgres
	MetricsRepo    repository.ObservationRepository // aggregated metrics (Postgres)
	ZDetector      *rules.ZScoreDetector
	Thresholds     *rules.Thresholds
	MLClient       *mlclient.Client
}

func (svc *ProcessService) generateID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

func NewProcessService(publisher repository.Publisher, alertRepo repository.AlertRepository, metricsRepo repository.ObservationRepository, mlClient *mlclient.Client) *ProcessService {
	return &ProcessService{
		AlertPublisher: publisher,
		AlertRepo:      alertRepo,
		MetricsRepo:    metricsRepo,
		ZDetector:      rules.NewZScoreDetector(30, 3.0, 0.1),
		Thresholds:     &rules.Thresholds{HeartRateMax: 100, SpO2Min: 90},
		MLClient:       mlClient,
	}
}

func (svc *ProcessService) HandleObservation(ctx context.Context, obs *entities.ObservationRecord) error {
	// 1. Check thresholds
	alertTypes, triggered := rules.CheckThresholds(obs, svc.Thresholds)
	if triggered {
		alertTypeStr := strings.Join(alertTypes, ", ")

		alert := entities.Alert{
			ID:            svc.generateID(),
			PatientID:     obs.PatientID,
			ObservationID: obs.ID,
			Type:          alertTypeStr,
			Message:       fmt.Sprintf("Alerts: %s at %s", alertTypeStr, obs.EffectiveDateTime),
			Timestamp:     time.Now(),
		}
		if err := svc.publishAndSaveAlert(ctx, &alert); err != nil {
			return fmt.Errorf("failed to handle threshold alert for patient %s: %v", alert.PatientID, err)
		}
	}

	// 2. Z-Score detection
	zAnomaly := svc.ZDetector.Add(obs.Value)

	var (
		mlAnomaly bool
		mlScore   float64
	)

	if svc.MLClient != nil {
		mlObs := mlclient.MLObservation{
			ID:                obs.ID,
			PatientID:         obs.PatientID,
			HeartRate:         obs.Value,
			RespRate:          0,
			Spo2:              0,
			EffectiveDateTime: obs.EffectiveDateTime.Format(time.RFC3339),
		}

		resp, err := svc.MLClient.Predict(mlObs)
		if err != nil {
			log.Printf("ML service error for obs %s: %v", obs.ID, err)
		} else if resp.Prediction {
			mlAnomaly = true
			mlScore = resp.AnomalyScore
		}
	}

	// 4. Alert if any anomaly is detected
	if zAnomaly || mlAnomaly {
		source := "Z-Score"
		if zAnomaly && mlAnomaly {
			source = "Z-Score and ML"
		} else if mlAnomaly {
			source = "ML"
		}

		message := fmt.Sprintf("Anomaly detected by %s: value=%.2f at %s", source, obs.Value, obs.EffectiveDateTime)
		if mlAnomaly {
			message += fmt.Sprintf(", ML anomaly score %.2f", mlScore)
		}

		alert := entities.Alert{
			ID:            svc.generateID(),
			PatientID:     obs.PatientID,
			ObservationID: obs.ID,
			Type:          "Anomaly",
			Message:       message,
			Timestamp:     time.Now(),
		}

		log.Printf("Anomaly alert for patient %s detected by %s", obs.PatientID, source)

		if err := svc.publishAndSaveAlert(ctx, &alert); err != nil {
			return fmt.Errorf("failed to handle anomaly alert: %v", err)
		}
	}

	// 5. Save metrics
	if err := svc.MetricsRepo.Save(ctx, obs); err != nil {
		return fmt.Errorf("failed to store metrics for patient %s: %v", obs.PatientID, err)
	}

	return nil
}

func (svc *ProcessService) publishAndSaveAlert(ctx context.Context, alert *entities.Alert) error {
	log.Printf("Publishing alert type %s for patient %s", alert.Type, alert.PatientID)
	if err := svc.AlertPublisher.PublishAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to publish alert: %v", err)
	}
	log.Printf("Alert published, saving in Postgres...")
	if err := svc.AlertRepo.Save(ctx, alert); err != nil {
		return fmt.Errorf("failed to save alert: %v", err)
	}
	log.Printf("Alert published in postgres with id %s", alert.ID)
	return nil
}
