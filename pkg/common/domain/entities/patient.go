package entities

import "time"

type Patient struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	BirthDate time.Time
}
