package repo

import "github.com/google/uuid"

type BaseModel interface {
	GetUUID() uuid.UUID
}
