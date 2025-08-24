package aria

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

type Conn struct {
	conn  *websocket.Conn
	store sync.Map
}

type Conns map[*Conn]struct{}

// Remove delete connection with argument connection.
func (c Conns) Remove(conn *Conn) {
	for cn := range c {
		if cn == conn {
			delete(c, cn)
		}
	}
}

// Slice convert map[*Conn]struct{} to []*Conn
func (c Conns) Slice() []*Conn {
	res := make([]*Conn, 0, len(c))
	for conn := range c {
		res = append(res, conn)
	}

	return res
}

func newConn(conn *websocket.Conn) *Conn {
	return &Conn{
		conn:  conn,
		store: sync.Map{},
	}
}

// Set can store any value.
func (c *Conn) Set(key string, value any) {
	c.store.Store(key, value)
}

// Get can retrive value that is set by Set method.
func (c *Conn) Get(key string) (value any, ok bool) {
	return c.store.Load(key)
}

// Delete can delete stored value that is set by Set method.
func (c *Conn) Delete(key string) {
	c.store.Delete(key)
}

// Ping can ping for the connection.
func (c *Conn) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}
