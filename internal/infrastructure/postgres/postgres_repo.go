package postgres

import (
	"context"
	"database/sql"
	"remote-patient-monitoring-system/internal/domain/model"

	_ "github.com/lib/pq"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(conn string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", conn)
	return &PostgresRepo{db: db}, err
}

func (r *PostgresRepo) Save(ctx context.Context, alert *model.Alert) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO alerts(id,patient_id,observation_id,message,type,timestamp,acknowledged)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		alert.ID, alert.PatientID, alert.ObservationID,
		alert.Message, alert.Type, alert.Timestamp, alert.Acknowledged)
	return err
}

func (r *PostgresRepo) FetchByPatient(ctx context.Context, patientID string) ([]model.Alert, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,observation_id,message,type,timestamp,acknowledged
		 FROM alerts WHERE patient_id=$1`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.ObservationID, &a.Message, &a.Type, &a.Timestamp, &a.Acknowledged); err != nil {
			continue
		}
		a.PatientID = patientID
		list = append(list, a)
	}
	return list, nil
}
