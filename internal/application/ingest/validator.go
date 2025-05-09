package ingest

import (
	"errors"
	"remote-patient-monitoring-system/internal/domain/model"

	v1 "github.com/robertoAraneda/go-fhir-validator/pkg/v1"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(obs *model.Observation) error {
	m, err := v1.ToStruct[map[string]interface{}](obs)
	if err != nil {
		return errors.New("failed to convert Observation to map: " + err.Error())
	}

	outcome, err := v1.ValidateResource(*m)
	if err != nil {
		return errors.New("FHIR validation error: " + err.Error())
	}

	if len(outcome.Issue) > 0 {
		return errors.New("FHIR validation failed: resource not conformant")
	}
	return nil
}
