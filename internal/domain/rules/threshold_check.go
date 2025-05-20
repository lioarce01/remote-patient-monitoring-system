package rules

import "remote-patient-monitoring-system/internal/domain/model"

// Thresholds definen the limits to send alerts
type Thresholds struct {
	HeartRateMax float64 `json:"heart_rate_max"`
	SpO2Min      float64 `json:"spo2_min"`
	// add more thresholds if needed
}

// CheckThresholds check an observation and send an alert
func CheckThresholds(obs *model.ObservationRecord, th *Thresholds) (alertType string, triggered bool) {
	switch obs.CodeText {
	case "heart-rate":
		// if heart rate exceeds the maximum
		if obs.Value > th.HeartRateMax {
			return "HighHeartRate", true
		}
	case "spo2":
		// if the oxygen saturation is lower than the minimum allowed
		if obs.Value < th.SpO2Min {
			return "LowSpO2", true
		}
	// example for corporal temperature
	// case "temperature":
	//     if obs.Value > th.TemperatureMax {
	//         return "HighTemperature", true
	//     }
	default:
		// if we dont have a rule, send false
		return "", false
	}

	return "", false
}
