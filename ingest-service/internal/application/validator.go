package application

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
	v1 "github.com/robertoAraneda/go-fhir-validator/pkg/v1"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(obs *entities.Observation) error {
	// serialize obs directly to json
	data, err := json.Marshal(obs)
	if err != nil {
		return fmt.Errorf("failed to marshal Observation: %w", err)
	}

	// protect of internal panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("FHIR validator panic: %v", r)
		}
	}()

	// validate using brute json
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
