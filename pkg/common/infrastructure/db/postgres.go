package db

import (
	"context"
	"log"

	_ "github.com/lib/pq"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
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

	// auto-migrate schema
	if err := db.AutoMigrate(&entities.Alert{}); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Save(ctx context.Context, alert *entities.Alert) error {
	log.Printf("[PostgresRepo] Trying to save alert: %+v", alert)
	err := r.db.WithContext(ctx).Create(alert).Error
	if err != nil {
		log.Printf("[PostgresRepo] Error saving alert: %v", err)
	} else {
		log.Printf("[PostgresRepo] Alerta saved successfully: %s", alert.ID)
	}
	return err
}

func (r *PostgresRepo) FetchByPatient(ctx context.Context, patientID string) ([]entities.Alert, error) {
	var alerts []entities.Alert
	err := r.db.WithContext(ctx).Where("patient_id = ?", patientID).Find(&alerts).Error
	return alerts, err
}
