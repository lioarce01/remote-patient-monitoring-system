package influxdb

import (
	"context"
	"fmt"
	"remote-patient-monitoring-system/internal/domain"
	"remote-patient-monitoring-system/internal/domain/model"

	client "github.com/influxdata/influxdb1-client/v2"
)

// InfluxRepo is a repository for storing observations in InfluxDB.
type InfluxRepo struct {
	client client.Client
	db     string
}

func NewInfluxRepo(addr, db, user, pass string) (domain.ObservationRepository, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: addr, Username: user, Password: pass})
	return &InfluxRepo{c, db}, err
}

func (r *InfluxRepo) Save(ctx context.Context, obs *model.Observation) error {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{Database: r.db, Precision: "s"})
	pt, _ := client.NewPoint("vitals",
		map[string]string{"patientID": obs.PatientID, "type": obs.Type},
		map[string]interface{}{"value": obs.Value}, obs.Timestamp)
	bp.AddPoint(pt)
	return r.client.Write(bp)
}

func (r *InfluxRepo) FetchObservations(ctx context.Context, patientID, from, to string) ([]model.Observation, error) {
	q := client.NewQuery(
		fmt.Sprintf(`SELECT * FROM vitals WHERE patientID='%s' AND time >= '%s' AND time <= '%s'`, patientID, from, to),
		r.db, "ns")
	resp, err := r.client.Query(q)
	if err != nil || resp.Error() != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	var out []model.Observation
	// mapear resp.Results a out...
	return out, nil
}
