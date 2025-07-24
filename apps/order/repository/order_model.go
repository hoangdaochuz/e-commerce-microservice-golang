package order_repository

import (
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Order struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

// func NewOrder(id uuid.UUID, name string) repo.Model {
// 	return &Order{
// 		ID:   id,
// 		Name: name,
// 	}
// }

func (o *Order) GetUUID() uuid.UUID {
	return o.ID
}
