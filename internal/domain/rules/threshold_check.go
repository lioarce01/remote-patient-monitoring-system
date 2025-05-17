package rules

import "remote-patient-monitoring-system/internal/domain/model"

// Thresholds define los límites para disparar alertas.
type Thresholds struct {
	HeartRateMax float64 `json:"heart_rate_max"`
	SpO2Min      float64 `json:"spo2_min"`
	// Añade aquí más umbrales según necesites, p.ej. TemperatureMax, BloodPressureMax, etc.
}

// CheckThresholds comprueba una observación y devuelve el tipo de alerta y un flag si se disparó.
func CheckThresholds(obs *model.ObservationRecord, th *Thresholds) (alertType string, triggered bool) {
	switch obs.CodeText {
	case "heart-rate":
		// Si la frecuencia cardíaca supera el máximo definido
		if obs.Value > th.HeartRateMax {
			return "HighHeartRate", true
		}
	case "spo2":
		// Si la saturación de oxígeno baja del mínimo definido
		if obs.Value < th.SpO2Min {
			return "LowSpO2", true
		}
	// Ejemplo para temperatura corporal, si la añades:
	// case "temperature":
	//     if obs.Value > th.TemperatureMax {
	//         return "HighTemperature", true
	//     }
	default:
		// Si no hay regla para este tipo, no disparamos alerta
		return "", false
	}

	return "", false
}
