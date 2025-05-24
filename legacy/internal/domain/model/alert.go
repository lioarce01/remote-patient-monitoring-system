package model

import "time"

type Alert struct {
	ID            string `gorm:"primaryKey"`
	PatientID     string `gorm:"index"`
	ObservationID string
	Message       string
	Type          string
	Timestamp     time.Time
	Acknowledged  bool
}
