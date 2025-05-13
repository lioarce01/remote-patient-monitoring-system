package influxdb

import (
	"context"
	"fmt"
	"remote-patient-monitoring-system/internal/domain"
	"remote-patient-monitoring-system/internal/domain/model"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

// InfluxRepo is a repository for storing observations in InfluxDB.
type InfluxRepo struct {
	client client.Client
	db     string
}

func NewInfluxRepo(addr, db, user, pass string) (domain.ObservationRepository, error) {
	if addr == "" || db == "" {
		return nil, fmt.Errorf("influxdb: addr and db must be provided")
	}
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: addr, Username: user, Password: pass,
	})
	if err != nil {
		return nil, fmt.Errorf("influxdb: client init failed: %w", err)
	}
	return &InfluxRepo{client: c, db: db}, nil
}

func (r *InfluxRepo) Save(ctx context.Context, record *model.ObservationRecord) error {
	// Asegúrate de que EffectiveDateTime es un time.Time
	timestamp := record.EffectiveDateTime

	// Crear el batch de puntos para InfluxDB
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{Database: r.db, Precision: "s"})

	// Crear el punto para la observación
	pt, _ := client.NewPoint(
		"vitals",
		map[string]string{
			"patientID": record.PatientID,
			"type":      record.Unit,
		},
		map[string]interface{}{
			"value": record.Value,
		},
		timestamp,
	)

	bp.AddPoint(pt)

	// Escribir el punto en InfluxDB
	return r.client.Write(bp)
}

func (r *InfluxRepo) FetchObservations(ctx context.Context, patientID, from, to string) ([]model.Observation, error) {
	q := client.NewQuery(
		fmt.Sprintf(`SELECT * FROM vitals WHERE patientID='%s' AND time >= '%s' AND time <= '%s'`, patientID, from, to),
		r.db, "ns")

	// Ejecutar la consulta
	resp, err := r.client.Query(q)
	if err != nil {
		return nil, fmt.Errorf("influx query failed: %w", err)
	}
	if resp.Error() != nil {
		return nil, fmt.Errorf("influx response error: %w", resp.Error())
	}

	var observations []model.Observation
	for _, result := range resp.Results {
		for _, series := range result.Series {
			for _, row := range series.Values {
				if len(row) < 4 {
					continue
				}

				timestampStr, ok := row[0].(string)
				if !ok {
					continue
				}
				timestamp, err := time.Parse(time.RFC3339, timestampStr)
				if err != nil {
					continue
				}

				valueFloat, ok := row[2].(float64)
				if !ok {
					continue
				}

				unitStr, ok := row[3].(string)
				if !ok {
					continue
				}

				record := model.ObservationRecord{
					PatientID:         patientID,
					Value:             valueFloat,
					Unit:              unitStr,
					EffectiveDateTime: timestamp,
				}

				obs := model.Observation{
					ID:                record.ID,
					ResourceType:      "Observation",
					Status:            "final",
					Code:              model.Code{Text: "Vital Sign"},
					Subject:           model.Subject{Reference: record.PatientID},
					EffectiveDateTime: record.EffectiveDateTime.Format(time.RFC3339),
					ValueQuantity: model.ValueQuantity{
						Value: record.Value,
						Unit:  record.Unit,
					},
				}

				observations = append(observations, obs)
			}
		}
	}

	return observations, nil
}
