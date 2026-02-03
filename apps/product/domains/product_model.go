package product_domains

import "github.com/google/uuid"

type Product struct {
	ID    *uuid.UUID `bson:"_id"`
	Name  string     `bson:"name"`
	Price int32      `bson:"price"`
}
