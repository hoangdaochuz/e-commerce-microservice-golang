//go:build integration

package order

import (
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	order_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/services/order"
	mocks "github.com/hoangdaochuz/ecommerce-microservice-golang/mocks/order_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// setupOrderGRPCServer sets up a gRPC server with bufconn for testing
func setupOrderGRPCServer(t *testing.T, mockRepo order_repository.OrderRepositoryInterface) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()

	// Create service with mock repository
	orderSvc := order_service.NewOrderServiceWithRepo(mockRepo)
	orderApp := NewOrderServiceApp(orderSvc)

	order.RegisterOrderServiceServer(server, orderApp)

	go func() {
		if err := server.Serve(lis); err != nil {
			// Server stopped, expected during test cleanup
		}
	}()

	t.Cleanup(func() {
		server.Stop()
		lis.Close()
	})

	return server, lis
}

// dialBufconn creates a gRPC client connection to the bufconn listener
func dialBufconn(ctx context.Context, lis *bufconn.Listener) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func TestOrderService_Integration_CreateOrder(t *testing.T) {
	// Use generated mock
	mockRepo := mocks.NewMockOrderRepositoryInterface(t)
	_, lis := setupOrderGRPCServer(t, mockRepo)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := order.NewOrderServiceClient(conn)

	t.Run("creates order via gRPC", func(t *testing.T) {
		resp, err := client.CreateOrder(ctx, &order.CreateOrderRequest{
			CustomerId: "customer_integration_test",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Contains(t, resp.OrderId, "order_customer_integration_test_")
	})

	t.Run("returns error for empty customer_id via gRPC", func(t *testing.T) {
		resp, err := client.CreateOrder(ctx, &order.CreateOrderRequest{
			CustomerId: "",
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "customer_id is required")
	})
}

func TestOrderService_Integration_GetOrderById(t *testing.T) {
	// Use generated mock
	mockRepo := mocks.NewMockOrderRepositoryInterface(t)
	_, lis := setupOrderGRPCServer(t, mockRepo)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := order.NewOrderServiceClient(conn)

	t.Run("gets order by id via gRPC", func(t *testing.T) {
		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		expectedOrder := &order_repository.Order{
			ID:   orderID,
			Name: "Integration Test Order",
		}

		mockRepo.EXPECT().
			FindOrderById(mock.Anything, orderID).
			Return(expectedOrder, nil).
			Once()

		resp, err := client.GetOrderById(ctx, &order.GetOrderByIdRequest{
			Id: orderID.String(),
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, orderID.String(), resp.Id)
		assert.Equal(t, "Integration Test Order", resp.Name)
	})

	t.Run("returns nil for non-existent order via gRPC", func(t *testing.T) {
		orderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

		mockRepo.EXPECT().
			FindOrderById(mock.Anything, orderID).
			Return(nil, nil).
			Once()

		resp, err := client.GetOrderById(ctx, &order.GetOrderByIdRequest{
			Id: orderID.String(),
		})

		// When order is not found, the handler returns nil, nil
		// gRPC will return an empty response
		require.NoError(t, err)
		assert.Nil(t, resp)
	})
}

func TestOrderService_Integration_ConcurrentRequests(t *testing.T) {
	// Use generated mock
	mockRepo := mocks.NewMockOrderRepositoryInterface(t)
	_, lis := setupOrderGRPCServer(t, mockRepo)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := order.NewOrderServiceClient(conn)

	t.Run("handles concurrent create order requests", func(t *testing.T) {
		concurrency := 10
		results := make(chan *order.CreateOrderResponse, concurrency)
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(idx int) {
				resp, err := client.CreateOrder(ctx, &order.CreateOrderRequest{
					CustomerId: "concurrent_customer",
				})
				if err != nil {
					errors <- err
				} else {
					results <- resp
				}
			}(i)
		}

		// Collect results
		successCount := 0
		errorCount := 0

		for i := 0; i < concurrency; i++ {
			select {
			case <-results:
				successCount++
			case <-errors:
				errorCount++
			}
		}

		assert.Equal(t, concurrency, successCount)
		assert.Equal(t, 0, errorCount)
	})
}
