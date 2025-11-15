package custom_nats

import (
	"context"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	"github.com/nats-io/nats.go"
)

// type (
// 	Handler interface{}
// )

type Connection struct {
	conn *nats.Conn
}

func NewConnection(conn *nats.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

type NatsConnWithCircuitBreaker struct {
	conn    *nats.Conn
	breaker *circuitbreaker.Breaker[*nats.Msg]
}

func NewNatsConnWithCircuitBreaker(conn *nats.Conn, breaker *circuitbreaker.Breaker[*nats.Msg]) *NatsConnWithCircuitBreaker {
	return &NatsConnWithCircuitBreaker{
		conn:    conn,
		breaker: breaker,
	}
}

type NatsSendRequest struct {
	Subject string
	Content []byte
}

func (ncc *NatsConnWithCircuitBreaker) SendRequest(ctx context.Context, req *NatsSendRequest) (*nats.Msg, error) {
	res, err := ncc.breaker.Do(ctx, func() (*nats.Msg, error) {
		return ncc.conn.RequestWithContext(ctx, req.Subject, req.Content)
	})
	return *res, err
}

func (ncc *NatsConnWithCircuitBreaker) GetNatsConn() *nats.Conn {
	return ncc.conn
}
