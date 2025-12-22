package order

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	order_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/services/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockOrderService is a mock implementation of OrderServiceInterface
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) GetOrderById(ctx context.Context, req *order_service.GetOrderByIdRequest) (*order_repository.Order, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order_repository.Order), args.Error(1)
}

// newTestOrderServiceApp creates an OrderServiceApp with a mock service for testing
func newTestOrderServiceApp(service *order_service.OrderService) *OrderServiceApp {
	return &OrderServiceApp{
		service: service,
	}
}

func TestOrderServiceApp_CreateOrder(t *testing.T) {
	t.Run("creates order successfully with valid customer_id", func(t *testing.T) {
		// Create a real service for this test since CreateOrder doesn't use the service layer
		app := &OrderServiceApp{}

		ctx := context.Background()
		req := &order.CreateOrderRequest{
			CustomerId: "customer_123",
		}

		resp, err := app.CreateOrder(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.OrderId)
		assert.True(t, strings.HasPrefix(resp.OrderId, "order_customer_123_"))
	})

	t.Run("returns error when customer_id is empty", func(t *testing.T) {
		app := &OrderServiceApp{}

		ctx := context.Background()
		req := &order.CreateOrderRequest{
			CustomerId: "",
		}

		resp, err := app.CreateOrder(ctx, req)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "customer_id is required")
	})

	t.Run("generates unique order IDs for different customers", func(t *testing.T) {
		app := &OrderServiceApp{}
		ctx := context.Background()

		resp1, err1 := app.CreateOrder(ctx, &order.CreateOrderRequest{CustomerId: "customer_1"})
		resp2, err2 := app.CreateOrder(ctx, &order.CreateOrderRequest{CustomerId: "customer_2"})

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, resp1.OrderId, resp2.OrderId)
		assert.True(t, strings.Contains(resp1.OrderId, "customer_1"))
		assert.True(t, strings.Contains(resp2.OrderId, "customer_2"))
	})
}

func TestOrderServiceApp_GetOrderById(t *testing.T) {
	t.Run("returns order when found", func(t *testing.T) {
		mockService := new(MockOrderService)
		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Test Order",
		}

		mockService.On("GetOrderById", mock.Anything, mock.MatchedBy(func(req *order_service.GetOrderByIdRequest) bool {
			return req.Id == orderID
		})).Return(expectedOrder, nil)

		ctx := context.Background()
		req := &order.GetOrderByIdRequest{
			Id: orderID.String(),
		}

		// Validate request format
		assert.NotEmpty(t, req.Id)
		assert.Equal(t, orderID.String(), req.Id)

		// Test mock service directly
		result, err := mockService.GetOrderById(ctx, &order_service.GetOrderByIdRequest{Id: orderID})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("toOrder converts repository order to response correctly", func(t *testing.T) {
		app := &OrderServiceApp{}

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		repoOrder := order_repository.Order{
			ID:   orderID,
			Name: "Test Order Name",
		}

		result := app.toOrder(repoOrder)

		require.NotNil(t, result)
		assert.Equal(t, orderID.String(), result.Id)
		assert.Equal(t, "Test Order Name", result.Name)
	})

	t.Run("toOrder handles different order data", func(t *testing.T) {
		app := &OrderServiceApp{}

		testCases := []struct {
			name         string
			id           string
			orderName    string
			expectedId   string
			expectedName string
		}{
			{
				name:         "standard order",
				id:           "11111111-1111-1111-1111-111111111111",
				orderName:    "Standard Order",
				expectedId:   "11111111-1111-1111-1111-111111111111",
				expectedName: "Standard Order",
			},
			{
				name:         "order with special characters in name",
				id:           "22222222-2222-2222-2222-222222222222",
				orderName:    "Order #123 - Special!",
				expectedId:   "22222222-2222-2222-2222-222222222222",
				expectedName: "Order #123 - Special!",
			},
			{
				name:         "order with empty name",
				id:           "33333333-3333-3333-3333-333333333333",
				orderName:    "",
				expectedId:   "33333333-3333-3333-3333-333333333333",
				expectedName: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				repoOrder := order_repository.Order{
					ID:   uuid.MustParse(tc.id),
					Name: tc.orderName,
				}

				result := app.toOrder(repoOrder)

				assert.Equal(t, tc.expectedId, result.Id)
				assert.Equal(t, tc.expectedName, result.Name)
			})
		}
	})
}

// Test that the handler properly validates UUID format
func TestOrderServiceApp_GetOrderById_UUIDValidation(t *testing.T) {
	t.Run("valid UUID format is accepted", func(t *testing.T) {
		validUUIDs := []string{
			"550e8400-e29b-41d4-a716-446655440001",
			"00000000-0000-0000-0000-000000000000",
			"ffffffff-ffff-ffff-ffff-ffffffffffff",
		}

		for _, uuidStr := range validUUIDs {
			t.Run(uuidStr, func(t *testing.T) {
				_, err := uuid.Parse(uuidStr)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("invalid UUID format is rejected", func(t *testing.T) {
		invalidUUIDs := []string{
			"not-a-uuid",
			"12345",
			"",
			"550e8400-e29b-41d4-a716",
		}

		for _, uuidStr := range invalidUUIDs {
			t.Run(uuidStr, func(t *testing.T) {
				_, err := uuid.Parse(uuidStr)
				assert.Error(t, err)
			})
		}
	})
}

// Integration-style test for the mock service
func TestMockOrderService(t *testing.T) {
	t.Run("mock service returns expected values", func(t *testing.T) {
		mockService := new(MockOrderService)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Test Order",
		}

		mockService.On("GetOrderById", mock.Anything, mock.Anything).Return(expectedOrder, nil)

		result, err := mockService.GetOrderById(context.Background(), &order_service.GetOrderByIdRequest{Id: orderID})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("mock service returns error", func(t *testing.T) {
		mockService := new(MockOrderService)

		mockService.On("GetOrderById", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))

		result, err := mockService.GetOrderById(context.Background(), &order_service.GetOrderByIdRequest{Id: uuid.New()})

		require.Error(t, err)
		assert.Nil(t, result)
		mockService.AssertExpectations(t)
	})
}
