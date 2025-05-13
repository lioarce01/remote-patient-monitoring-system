package ingest

import (
	"remote-patient-monitoring-system/internal/domain/model"
	"time"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) FromTelemetry(in TelemetryInput) *model.Observation {
	return &model.Observation{
		ResourceType:      "Observation",                                       // Required by FHIR
		Status:            "final",                                             // You can change the status based on your logic
		Code:              model.Code{Text: "Heart rate"},                      // Type of observation
		Subject:           model.Subject{Reference: "Patient/" + in.PatientID}, // Reference to patient
		EffectiveDateTime: in.Timestamp.Format(time.RFC3339),                   // RFC3339 format for date
		ValueQuantity: model.ValueQuantity{
			Value: in.HeartRate,
			Unit:  "bpm", // Unit of measurement
		},
	}
}
