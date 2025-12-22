// Package testutil provides gRPC testing utilities
package testutil

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// GRPCTestServer wraps a gRPC server for testing using in-memory bufconn
type GRPCTestServer struct {
	Server   *grpc.Server
	Listener *bufconn.Listener
	t        *testing.T
}

// NewGRPCTestServer creates a new gRPC test server with bufconn
func NewGRPCTestServer(t *testing.T) *GRPCTestServer {
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()

	ts := &GRPCTestServer{
		Server:   server,
		Listener: lis,
		t:        t,
	}

	return ts
}

// Start starts the gRPC test server in a goroutine
func (s *GRPCTestServer) Start() {
	go func() {
		if err := s.Server.Serve(s.Listener); err != nil {
			// Server stopped, this is expected during shutdown
		}
	}()

	s.t.Cleanup(func() {
		s.Server.Stop()
	})
}

// Dial creates a client connection to the test server
func (s *GRPCTestServer) Dial(ctx context.Context) *grpc.ClientConn {
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return s.Listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		s.t.Fatalf("Failed to dial bufnet: %v", err)
	}

	s.t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// GRPCClientConn creates a gRPC client connection for testing
func GRPCClientConn(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

// WithGRPCServer is a helper that creates a gRPC server, registers services, and returns a client connection
func WithGRPCServer(t *testing.T, registerFunc func(*grpc.Server)) *grpc.ClientConn {
	ts := NewGRPCTestServer(t)
	registerFunc(ts.Server)
	ts.Start()
	return ts.Dial(context.Background())
}
