package model

import "time"

type Patient struct {
	ID        string
	Name      string
	BirthDate time.Time
	// Other..
}
