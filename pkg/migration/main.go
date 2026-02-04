package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
	"go.mongodb.org/mongo-driver/bson"
)

type Migrationer interface {
	Name() string
	Up(ctx context.Context, dbClient repo.IDBConnection) error
	Down(ctx context.Context, dbClient repo.IDBConnection) error
}

type MigrationRunner struct {
	ctx          context.Context
	client       repo.IDBClient
	migrationers []Migrationer
}

func NewMigrationRunner(ctx context.Context, client repo.IDBClient) *MigrationRunner {
	return &MigrationRunner{
		ctx:          ctx,
		client:       client,
		migrationers: make([]Migrationer, 0),
	}
}

func (m *MigrationRunner) Register(migrationers ...Migrationer) error {
	m.migrationers = append(m.migrationers, migrationers...)
	return nil
}

type MigrationRunnerType string

const (
	UP   MigrationRunnerType = "UP"
	DOWN MigrationRunnerType = "DOWN"
	BOTH MigrationRunnerType = "BOTH"
)

func (m *MigrationRunner) Run(runnerType MigrationRunnerType) error {
	if len(m.migrationers) == 0 {
		return fmt.Errorf("There are not any script for migration")
	}
	for _, migrationer := range m.migrationers {
		var out MigrationModel
		migrationName := migrationer.Name()
		// Check whether this migrationer already migrated
		err := m.client.FindMigrationerByName(m.ctx, &out, bson.M{"name": migrationName})
		if err != nil {
			return fmt.Errorf("Migrationer %s fail: %w", migrationName, err)
		}
		if out.Id != nil {
			logging.GetSugaredLogger().Infof("Migration %s has already migrated. SKIP", migrationName)
			continue
		}
		err = m.runMigration(m.ctx, runnerType, migrationer)
		if err != nil {
			return fmt.Errorf("fail to migrate migrationer: %s", migrationName)
		}
		id := uuid.New()
		migrated := MigrationModel{
			Id:          &id,
			Name:        migrationName,
			MigrationAt: time.Now(),
		}
		return m.client.Insert(m.ctx, migrated)
	}
	return nil
}

func (m *MigrationRunner) runMigration(ctx context.Context, runnerType MigrationRunnerType, migrationer Migrationer) error {
	switch runnerType {
	case UP:
		return migrationer.Up(ctx, m.client.GetConnection())
	case DOWN:
		return migrationer.Down(ctx, m.client.GetConnection())
	case BOTH:
		err := migrationer.Up(ctx, m.client.GetConnection())
		if err != nil {
			return err
		}
		return migrationer.Down(ctx, m.client.GetConnection())
	default:
		return fmt.Errorf("No match migration runner type")
	}
}
