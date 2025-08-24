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

func newConn(conn *websocket.Conn) *Conn {
	return &Conn{
		conn:  conn,
		store: sync.Map{},
	}
}

func (c *Conn) Set(key string, value any) {
	c.store.Store(key, value)
}

func (c *Conn) Get(key string) (value any, ok bool) {
	return c.store.Load(key)
}

func (c *Conn) Delete(key string) {
	c.store.Delete(key)
}

func (c *Conn) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}
