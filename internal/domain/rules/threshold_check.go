package rules

import "remote-patient-monitoring-system/internal/domain/model"

type Thresholds struct {
	HeartRateMax float64 `json:"heart_rate_max"`
	SpO2Min      float64 `json:"spo2_min"`
	//...
}

func CheckThresholds(obs *model.ObservationRecord, th *Thresholds) (alertType string, triggered bool) {
	switch obs.CodeText { // Accede al tipo de observación a través de Code.Text
	case "heart-rate":
		if obs.Value > th.HeartRateMax { // Accede al valor a través de ValueQuantity.Value
			return "HighHeartRate", true
		}
	case "spo2":
		if obs.Value < th.SpO2Min { // Similar para SpO2
			return "LowSpO2", true
		}
		// Agregar más casos para otros tipos de observaciones
	}

	return "", false
}
