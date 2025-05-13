package postgres

import (
	"context"
	"remote-patient-monitoring-system/internal/domain/model"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresRepo struct {
	db *gorm.DB
}

func NewPostgresRepo(conn string) (*PostgresRepo, error) {
	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrar el esquema
	if err := db.AutoMigrate(&model.Alert{}); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Save(ctx context.Context, alert *model.Alert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *PostgresRepo) FetchByPatient(ctx context.Context, patientID string) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).Find(&alerts).Error
	return alerts, err
}
