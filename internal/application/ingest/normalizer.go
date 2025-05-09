package ingest

import (
	"remote-patient-monitoring-system/internal/domain/model"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) FromTelemetry(in TelemetryInput) *model.Observation {
	return &model.Observation{
		ID:        in.PatientID + "-" + in.Timestamp.Format("20060102150405"),
		PatientID: in.PatientID,
		Type:      "heart-rate",
		Value:     in.HeartRate,
		Unit:      "bpm",
		Timestamp: in.Timestamp,
	}
}
