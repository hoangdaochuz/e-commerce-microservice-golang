package order_service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	mocks "github.com/hoangdaochuz/ecommerce-microservice-golang/mocks/order_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderService_GetOrderById(t *testing.T) {
	t.Run("returns order when found", func(t *testing.T) {
		// Use generated mock - auto cleanup and assert expectations
		mockRepo := mocks.NewMockOrderRepositoryInterface(t)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Test Order",
		}

		// Use EXPECT() pattern from generated mock
		mockRepo.EXPECT().
			FindOrderById(mock.Anything, orderID).
			Return(expectedOrder, nil)

		service := NewOrderServiceWithRepo(mockRepo)

		ctx := context.Background()
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		assert.Equal(t, "Test Order", result.Name)
		// AssertExpectations is called automatically by t.Cleanup()
	})

	t.Run("returns nil when order not found (sql.ErrNoRows)", func(t *testing.T) {
		mockRepo := mocks.NewMockOrderRepositoryInterface(t)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

		mockRepo.EXPECT().
			FindOrderById(mock.Anything, orderID).
			Return(nil, sql.ErrNoRows)

		service := NewOrderServiceWithRepo(mockRepo)

		ctx := context.Background()
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockRepo := mocks.NewMockOrderRepositoryInterface(t)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
		expectedErr := errors.New("database connection failed")

		mockRepo.EXPECT().
			FindOrderById(mock.Anything, orderID).
			Return(nil, expectedErr)

		service := NewOrderServiceWithRepo(mockRepo)

		ctx := context.Background()
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("handles different order IDs correctly", func(t *testing.T) {
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
				mockRepo := mocks.NewMockOrderRepositoryInterface(t)

				expectedOrder := &order_repository.Order{
					ID:   tc.orderID,
					Name: tc.orderName,
				}

				mockRepo.EXPECT().
					FindOrderById(mock.Anything, tc.orderID).
					Return(expectedOrder, nil)

				service := NewOrderServiceWithRepo(mockRepo)

				ctx := context.Background()
				req := &GetOrderByIdRequest{Id: tc.orderID}

				result, err := service.GetOrderById(ctx, req)

				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.orderID, result.ID)
				assert.Equal(t, tc.orderName, result.Name)
			})
		}
	})

	t.Run("passes context correctly to repository", func(t *testing.T) {
		mockRepo := mocks.NewMockOrderRepositoryInterface(t)

		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Context Test Order",
		}

		// Use a specific context to verify it's passed through
		type ctxKey string
		ctx := context.WithValue(context.Background(), ctxKey("test"), "value")

		mockRepo.EXPECT().
			FindOrderById(ctx, orderID).
			Return(expectedOrder, nil)

		service := NewOrderServiceWithRepo(mockRepo)
		req := &GetOrderByIdRequest{Id: orderID}

		result, err := service.GetOrderById(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestNewOrderServiceWithRepo(t *testing.T) {
	t.Run("creates service with provided repository", func(t *testing.T) {
		mockRepo := mocks.NewMockOrderRepositoryInterface(t)
		service := NewOrderServiceWithRepo(mockRepo)

		require.NotNil(t, service)
		assert.Equal(t, mockRepo, service.OrderRepo)
	})
}
