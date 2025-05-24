package application

import (
	"log"
	"time"

	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) FromTelemetry(input TelemetryInput) *entities.Observation {
	log.Printf("[Normalizer] Creating observation of type: %s with value: %f", input.Type, input.Value)

	return &entities.Observation{
		ResourceType:      "Observation",
		Status:            "Final",
		Code:              entities.Code{Text: input.Type},
		Subject:           entities.Subject{Reference: input.PatientID},
		EffectiveDateTime: input.Timestamp.Format(time.RFC3339),
		ValueQuantity: entities.ValueQuantity{
			Value: input.Value,
			Unit:  input.Unit,
		},
	}
}
