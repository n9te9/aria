package aria

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

type aria struct {
	onPong          func(ctx context.Context, conn *Conn) error
	onConnect       func(ctx context.Context, conn *Conn) error
	onDisconnect    func(ctx context.Context, conn *Conn) error
	onClose         func(ctx context.Context, conn *Conn) error
	onMessage       func(ctx context.Context, conn *Conn, message []byte) error
	onMessageBinary func(ctx context.Context, conn *Conn, message []byte) error
	onError         func(ctx context.Context, conn *Conn, err error)

	insecureSkipVerify   bool
	subprotocols         []string
	originPatterns       []string
	compressionMode      websocket.CompressionMode
	compressionThreshold int
	onPingReceived       func(ctx context.Context, payload []byte) bool
	onPongReceived       func(ctx context.Context, payload []byte)

	conns []*Conn
}

type FilterFunc func(conns []*Conn) bool

type Aria interface {
	BroadCast(ctx context.Context, message []byte) error
	BroadCastFilter(ctx context.Context, message []byte, filter FilterFunc) error

	Handle(w http.ResponseWriter, r *http.Request) error

	OnConnect(func(ctx context.Context, conn *Conn) error)
	OnPong(func(ctx context.Context, conn *Conn) error)
	OnMessage(func(ctx context.Context, conn *Conn, message []byte) error)
	OnMessageBinary(func(ctx context.Context, conn *Conn, message []byte) error)
	OnDisconnect(func(ctx context.Context, conn *Conn) error)
	OnClose(func(ctx context.Context, conn *Conn) error)
	OnError(func(ctx context.Context, conn *Conn, err error))
}

var _ Aria = (*aria)(nil)

func New(opts ...Option) *aria {
	a := &aria{}

	for _, opt := range opts {
		opt.Apply(a)
	}
	return a
}

func (a *aria) Handle(w http.ResponseWriter, r *http.Request) error {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify:   a.insecureSkipVerify,
		Subprotocols:         a.subprotocols,
		OriginPatterns:       a.originPatterns,
		CompressionMode:      a.compressionMode,
		CompressionThreshold: a.compressionThreshold,
		OnPingReceived:       a.onPingReceived,
		OnPongReceived:       a.onPongReceived,
	})
	if err != nil {
		return err
	}

	ariaConn := newConn(conn)
	a.conns = append(a.conns, ariaConn)

	ctx := r.Context()
	for _, conn := range a.conns {
		go func(c *Conn) {
			if err := a.run(ctx, c); err != nil {
				a.onError(ctx, c, err)
			}
		}(conn)
	}

	return nil
}

func (a *aria) run(ctx context.Context, c *Conn) (e error) {
	for {
		typ, msg, err := c.conn.Read(ctx)
		if err != nil {
			e = err
			return
		}

		switch typ {
		case websocket.MessageText:
			if err := a.onMessage(ctx, c, msg); err != nil {
				e = err
				return
			}
		case websocket.MessageBinary:
			if err := a.onMessageBinary(ctx, c, msg); err != nil {
				e = err
				return
			}
		}
	}
}

func (a *aria) BroadCast(ctx context.Context, message []byte) error {
	for _, conn := range a.conns {
		if err := conn.conn.Write(ctx, websocket.MessageText, message); err != nil {
			return err
		}
	}

	return nil
}

func (a *aria) BroadCastFilter(ctx context.Context, message []byte, filter FilterFunc) error {
	for _, conn := range a.conns {
		if filter(a.conns) {
			if err := conn.conn.Write(ctx, websocket.MessageText, message); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *aria) OnPong(fn func(ctx context.Context, conn *Conn) error) {
	a.onPong = fn
}

func (a *aria) OnConnect(fn func(ctx context.Context, conn *Conn) error) {
	a.onConnect = fn
}

func (a *aria) OnMessage(fn func(ctx context.Context, conn *Conn, message []byte) error) {
	a.onMessage = fn
}

func (a *aria) OnClose(fn func(ctx context.Context, conn *Conn) error) {
	a.onClose = fn
}

func (a *aria) OnDisconnect(fn func(ctx context.Context, conn *Conn) error) {
	a.onDisconnect = fn
}

func (a *aria) OnMessageBinary(fn func(ctx context.Context, conn *Conn, message []byte) error) {
	a.onMessageBinary = fn
}

func (a *aria) OnError(fn func(ctx context.Context, conn *Conn, err error)) {
	a.onError = fn
}
