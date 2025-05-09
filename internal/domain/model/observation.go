package model

import "time"

type Observation struct {
	ID        string    `json:"id,omitempty"`
	PatientID string    `json:"subject"`           // fhir uses "subject" for patient reference
	Type      string    `json:"code"`              // "heart-rate", "blood-pressure", etc
	Value     float64   `json:"valueQuantity"`     // numeric measurement
	Unit      string    `json:"unit"`              // "bpm", "mmHg", "%", etc
	Timestamp time.Time `json:"effectiveDateTime"` // when measurement was taken
}
