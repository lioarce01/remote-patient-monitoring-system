package influxdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	timestamp := record.EffectiveDateTime

	// create a batch for influxdb
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{Database: r.db, Precision: "s"})

	// create the point for the obs
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

	log.Printf("[InfluxRepo] Saving value: %f", record.Value)

	// write the point into influxdb
	return r.client.Write(bp)
}

func (r *InfluxRepo) FetchObservations(ctx context.Context, patientID, from, to string) ([]model.Observation, error) {
	query := fmt.Sprintf(`
		SELECT * FROM vitals 
		WHERE patientID = '%s' 
		AND time >= '%s' 
		AND time <= '%s'
	`, patientID, from, to)

	log.Printf("Generated InfluxQL query: %s", query)

	q := client.NewQuery(query, r.db, "s")

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

				var timestamp time.Time
				switch v := row[0].(type) {
				case string:
					t, err := time.Parse(time.RFC3339, v)
					if err != nil {
						log.Printf("[FetchObservations] invalid timestamp format: %v", err)
						continue
					}
					timestamp = t
				case json.Number:
					n, err := v.Int64()
					if err != nil {
						log.Printf("[FetchObservations] invalid timestamp number: %v", err)
						continue
					}
					// convert nanoseconds unix to time.time
					timestamp = time.Unix(0, n)
				default:
					log.Printf("[FetchObservations] unexpected timestamp type: %T", v)
					continue
				}

				unitStr, ok := row[2].(string)
				if !ok {
					log.Printf("[FetchObservations] expected unit as string, got: %T", row[2])
					continue
				}

				var valueFloat float64
				switch v := row[3].(type) {
				case float64:
					valueFloat = v
				case json.Number:
					f, err := v.Float64()
					if err != nil {
						log.Printf("[FetchObservations] failed to convert value json.Number to float64: %v", err)
						continue
					}
					valueFloat = f
				default:
					log.Printf("[FetchObservations] unexpected type for value: %T", v)
					continue
				}

				obs := model.Observation{
					ResourceType:      "Observation",
					Status:            "Final",
					Code:              model.Code{Text: "Vital Sign"},
					Subject:           model.Subject{Reference: patientID},
					EffectiveDateTime: timestamp.Format(time.RFC3339),
					ValueQuantity: model.ValueQuantity{
						Value: valueFloat,
						Unit:  unitStr,
					},
				}

				observations = append(observations, obs)
			}
		}
	}

	return observations, nil
}
