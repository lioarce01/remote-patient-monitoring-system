package rules

import "remote-patient-monitoring-system/internal/domain/model"

// Thresholds definen the limits to send alerts
type Thresholds struct {
	HeartRateMax float64 `json:"heart_rate_max"`
	SpO2Min      float64 `json:"spo2_min"`
	// add more thresholds if needed
}

// CheckThresholds check an observation and send an alert
func CheckThresholds(obs *model.ObservationRecord, th *Thresholds) (alertTypes []string, triggered bool) {
	switch obs.CodeText {
	case "heart-rate":
		if obs.Value > th.HeartRateMax {
			alertTypes = append(alertTypes, "HighHeartRate")
		}
	case "spo2":
		if obs.Value < th.SpO2Min {
			alertTypes = append(alertTypes, "LowSpO2")
		}
		// Agrega más métricas
	}

	if len(alertTypes) > 0 {
		return alertTypes, true
	}
	return nil, false
}
