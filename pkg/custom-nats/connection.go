package custom_nats

import (
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

// func (c *Connection) Subcribe(subject string, handler Handler) error {
// 	return nil
// }
