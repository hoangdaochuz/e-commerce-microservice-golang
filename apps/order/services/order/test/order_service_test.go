package order_service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	order_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/services/order"
	orderRepoMock "github.com/hoangdaochuz/ecommerce-microservice-golang/mocks/order_repository"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_OrderService(t *testing.T) {
	orderRepo := orderRepoMock.NewMockOrderRepositoryInterface(t)
	t.Run("Test_GetOrderById_Success", func(t *testing.T) {
		orderId := uuid.MustParse("9f2c7a3e-6b8d-4e5f-9a1c-3d4f8b2e6a90")
		orderResultExpect := &order_repository.Order{
			ID:   orderId,
			Name: "Test order",
		}
		orderRepo.EXPECT().FindOrderById(mock.Anything, orderId).Return(orderResultExpect, nil)

		orderService := order_service.NewOrderService(orderRepo)
		res, err := orderService.GetOrderById(context.Background(), &order_service.GetOrderByIdRequest{
			Id: orderId,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, orderResultExpect.ID, res.ID)
		require.Equal(t, orderResultExpect.Name, res.Name)
	})

	t.Run("Test_GetOrderById_NotFound", func(t *testing.T) {
		orderId := uuid.MustParse("9f2c7a3e-6b8d-4e5f-9a1c-3d4f8b2e6a91")
		orderRepo.EXPECT().FindOrderById(mock.Anything, orderId).Return(nil, sql.ErrNoRows)

		orderService := order_service.NewOrderService(orderRepo)
		res, err := orderService.GetOrderById(context.Background(), &order_service.GetOrderByIdRequest{
			Id: orderId,
		})
		require.Nil(t, res)
		require.Nil(t, err)
	})

	t.Run("Test_CreateOrderWithTransaction_Success", func(t *testing.T) {
		// Implement later due to order repo has not implemented yet
		t.Skip()
	})

}
