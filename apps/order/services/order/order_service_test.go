package order_service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockOrderRepository is a mock implementation of OrderRepositoryInterface
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) FindOrderById(ctx context.Context, id uuid.UUID) (*order_repository.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order_repository.Order), args.Error(1)
}

func (m *MockOrderRepository) CreateOrderWithTransaction(ctx context.Context, data order_repository.Order, other ...interface{}) (interface{}, error) {
	args := m.Called(ctx, data, other)
	return args.Get(0), args.Error(1)
}

func TestOrderService_GetOrderById(t *testing.T) {
	t.Run("returns order when found", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		service := NewOrderServiceWithRepo(mockRepo)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Test Order",
		}

		mockRepo.On("FindOrderById", mock.Anything, orderID).Return(expectedOrder, nil)

		ctx := context.Background()
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		assert.Equal(t, "Test Order", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns nil when order not found (sql.ErrNoRows)", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		service := NewOrderServiceWithRepo(mockRepo)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

		mockRepo.On("FindOrderById", mock.Anything, orderID).Return(nil, sql.ErrNoRows)

		ctx := context.Background()
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		service := NewOrderServiceWithRepo(mockRepo)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
		expectedErr := errors.New("database connection failed")

		mockRepo.On("FindOrderById", mock.Anything, orderID).Return(nil, expectedErr)

		ctx := context.Background()
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles different order IDs correctly", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		service := NewOrderServiceWithRepo(mockRepo)

		testCases := []struct {
			name      string
			orderID   uuid.UUID
			orderName string
		}{
			{"first order", uuid.MustParse("11111111-1111-1111-1111-111111111111"), "Order 1"},
			{"second order", uuid.MustParse("22222222-2222-2222-2222-222222222222"), "Order 2"},
			{"third order", uuid.MustParse("33333333-3333-3333-3333-333333333333"), "Order 3"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				expectedOrder := &order_repository.Order{
					ID:   tc.orderID,
					Name: tc.orderName,
				}
				mockRepo.On("FindOrderById", mock.Anything, tc.orderID).Return(expectedOrder, nil).Once()

				ctx := context.Background()
				req := &GetOrderByIdRequest{Id: tc.orderID}

				result, err := service.GetOrderById(ctx, req)

				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.orderID, result.ID)
				assert.Equal(t, tc.orderName, result.Name)
			})
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("passes context correctly to repository", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		service := NewOrderServiceWithRepo(mockRepo)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Context Test Order",
		}

		// Use a specific context to verify it's passed through
		type ctxKey string
		ctx := context.WithValue(context.Background(), ctxKey("test"), "value")

		mockRepo.On("FindOrderById", ctx, orderID).Return(expectedOrder, nil)

		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNewOrderServiceWithRepo(t *testing.T) {
	t.Run("creates service with provided repository", func(t *testing.T) {
		mockRepo := new(MockOrderRepository)
		service := NewOrderServiceWithRepo(mockRepo)

		require.NotNil(t, service)
		assert.Equal(t, mockRepo, service.OrderRepo)
	})
}
