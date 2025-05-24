package ingest

import (
	"log"
	"remote-patient-monitoring-system/internal/domain/model"
	"time"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) FromTelemetry(input TelemetryInput) *model.Observation {
	log.Printf("[Normalizer] Creating observation of type: %s with value: %f", input.Type, input.Value)

	return &model.Observation{
		ResourceType:      "Observation",
		Status:            "Final",
		Code:              model.Code{Text: input.Type},
		Subject:           model.Subject{Reference: input.PatientID},
		EffectiveDateTime: input.Timestamp.Format(time.RFC3339),
		ValueQuantity: model.ValueQuantity{
			Value: input.Value,
			Unit:  input.Unit,
		},
	}
}
