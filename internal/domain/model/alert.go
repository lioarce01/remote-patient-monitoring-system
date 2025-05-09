package model

import "time"

type Alert struct {
	ID            string
	PatientID     string
	ObservationID string    // reference to Observation.ID that triggered the alert
	Message       string    // human-readable alert message
	Type          string    // "warning" or "critical"
	Timestamp     time.Time // Wwen the alert was generated
	Acknowledged  bool      // wether the alert was acknowledged by a user
}
