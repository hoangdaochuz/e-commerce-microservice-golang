package migration

import (
	"time"

	"github.com/google/uuid"
)

type MigrationModel struct {
	Id          *uuid.UUID `bson:"_id"`
	Name        string     `bson:"name,omitempty"`
	MigrationAt time.Time  `bson:"migration_at,omitempty"`
}

func (m *MigrationModel) GetUUID() uuid.UUID {
	return *m.Id
}
