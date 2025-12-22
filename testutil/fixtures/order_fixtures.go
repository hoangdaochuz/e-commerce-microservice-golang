// Package fixtures provides test data fixtures for the e-commerce microservice tests
package fixtures

import (
	"time"

	"github.com/google/uuid"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
)

// Order fixtures for testing

// ValidOrderID returns a valid UUID for testing
func ValidOrderID() uuid.UUID {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
}

// ValidCustomerID returns a valid customer ID for testing
func ValidCustomerID() string {
	return "customer_123"
}

// SampleOrder returns a sample order for testing
func SampleOrder() *order_repository.Order {
	return &order_repository.Order{
		ID:   ValidOrderID(),
		Name: "Test Order",
	}
}

// SampleOrderWithID returns a sample order with a custom ID
func SampleOrderWithID(id uuid.UUID) *order_repository.Order {
	return &order_repository.Order{
		ID:   id,
		Name: "Test Order",
	}
}

// MultipleOrders returns multiple sample orders for testing
func MultipleOrders(count int) []*order_repository.Order {
	orders := make([]*order_repository.Order, count)
	for i := 0; i < count; i++ {
		orders[i] = &order_repository.Order{
			ID:   uuid.New(),
			Name: "Test Order " + string(rune('A'+i)),
		}
	}
	return orders
}

// ExpectedOrderID generates an expected order ID based on customer and timestamp
func ExpectedOrderID(customerID string, timestamp time.Time) string {
	return "order_" + customerID + "_" + string(rune(timestamp.Unix()))
}
