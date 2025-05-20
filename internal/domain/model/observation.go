package model

import (
	"errors"
	"log"
	"time"
)

type ObservationRecord struct {
	ID                string `gorm:"primaryKey"`
	ResourceType      string
	Status            string
	CodeText          string
	PatientID         string
	Subject           string
	EffectiveDateTime time.Time
	Value             float64
	Unit              string
}

type Observation struct {
	ID                string        `json:"id,omitempty"`
	ResourceType      string        `json:"resourceType"`
	Status            string        `json:"status"`
	Code              Code          `json:"code"`
	Subject           Subject       `json:"subject"`
	EffectiveDateTime string        `json:"effectiveDateTime"`
	ValueQuantity     ValueQuantity `json:"valueQuantity"`
}

type Code struct {
	Text string `json:"text"`
}

type Subject struct {
	Reference string `json:"reference"`
}

type ValueQuantity struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func ToObservationRecord(obs *Observation) (*ObservationRecord, error) {
	if obs == nil {
		return nil, errors.New("observation is nil")
	}

	// convert EffectiveDateTime string to time.time
	effectiveDateTime, err := time.Parse(time.RFC3339, obs.EffectiveDateTime)
	if err != nil {
		return nil, err
	}

	// create and return `ObservationRecord`.
	record := &ObservationRecord{
		ID:                obs.ID,
		ResourceType:      obs.ResourceType,
		Status:            obs.Status,
		CodeText:          obs.Code.Text,
		PatientID:         obs.Subject.Reference,
		Subject:           obs.Subject.Reference,
		EffectiveDateTime: effectiveDateTime,
		Value:             obs.ValueQuantity.Value,
		Unit:              obs.ValueQuantity.Unit,
	}

	log.Printf("[ToObservationRecord] Value: %f", record.Value)

	return record, nil
}
