package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/odpf/guardian/pkg/audit"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type auditPostgresModel struct {
	TraceID   string
	Timestamp time.Time
	Action    string
	Actor     string
	Data      datatypes.JSON
	App       datatypes.JSON
}

func (a auditPostgresModel) TableName() string {
	return "audit_logs"
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db}
}

func (r *PostgresRepository) Init(ctx context.Context) error {
	if err := r.db.WithContext(ctx).AutoMigrate(&auditPostgresModel{}); err != nil {
		return fmt.Errorf("migrating audit model to postgres db: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Insert(ctx context.Context, l *audit.Log) error {
	data, err := json.Marshal(l.Data)
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	app, err := json.Marshal(l.App)
	if err != nil {
		return fmt.Errorf("marshaling app: %w", err)
	}
	m := &auditPostgresModel{
		TraceID:   l.TraceID,
		Timestamp: l.Timestamp,
		Action:    l.Action,
		Actor:     l.Actor,
		Data:      datatypes.JSON(data),
		App:       datatypes.JSON(app),
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("inserting to db: %w", err)
	}

	return nil
}
