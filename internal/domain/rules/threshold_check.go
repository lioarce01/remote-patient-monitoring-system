package rules

import "remote-patient-monitoring-system/internal/domain/model"

type Thresholds struct {
	HeartRateMax float64
	SpO2Min      float64
	//...
}

func CheckThresholds(obs *model.Observation, th *Thresholds) (alertType string, triggered bool) {
	switch obs.Type {
	case "heart-rate":
		if obs.Value > th.HeartRateMax {
			return "HighHeartRate", true
		}
	case "spo2":
		if obs.Value < th.SpO2Min {
			return "LowSpO2", true
		}
		// Add more cases for other observation types
	}

	return "", false
}
