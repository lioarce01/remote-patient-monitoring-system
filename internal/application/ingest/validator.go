package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"remote-patient-monitoring-system/internal/domain/model"

	v1 "github.com/robertoAraneda/go-fhir-validator/pkg/v1"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(obs *model.Observation) error {
	// Serializar obs directamente a JSON
	data, err := json.Marshal(obs)
	if err != nil {
		return fmt.Errorf("failed to marshal Observation: %w", err)
	}

	// Protegernos de cualquier panic interno
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("FHIR validator panic: %v", r)
		}
	}()

	// Validar usando el JSON bruto
	var resource map[string]interface{}
	if err := json.Unmarshal(data, &resource); err != nil {
		return fmt.Errorf("failed to unmarshal Observation to map: %w", err)
	}
	outcome, err := v1.ValidateResource(resource)
	if err != nil {
		return fmt.Errorf("FHIR validation error: %w", err)
	}
	if len(outcome.Issue) > 0 {
		return errors.New("FHIR validation failed: resource not conformant")
	}
	return nil
}
