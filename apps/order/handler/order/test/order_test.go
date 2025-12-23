package order_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	order_api "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/handler/order"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	order_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/services/order"
	orderServiceMock "github.com/hoangdaochuz/ecommerce-microservice-golang/mocks/order_service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_OrderEndpointHandler(t *testing.T) {
	orderServiceMock := orderServiceMock.NewMockOrderServiceInterface(t)
	t.Run("Test_GetOrderById_Success", func(t *testing.T) {
		orderId := "9f2c7a3e-6b8d-4e5f-9a1c-3d4f8b2e6a90"
		orderExpect := &order_repository.Order{
			ID:   uuid.MustParse(orderId),
			Name: "Order test",
		}
		orderServiceMock.EXPECT().GetOrderById(mock.Anything, mock.MatchedBy(func(req *order_service.GetOrderByIdRequest) bool {
			return req.Id == uuid.MustParse(orderId)
		})).Return(orderExpect, nil).Once()

		orderApp := order.NewOrderServiceApp(orderServiceMock)
		res, err := orderApp.GetOrderById(context.Background(), &order_api.GetOrderByIdRequest{
			Id: orderId,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, orderId, res.Id)
		require.Equal(t, orderExpect.Name, res.Name)
	})
}
