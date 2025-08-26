package aria

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type aria struct {
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

	mutex *sync.Mutex
	conns Conns
}

type FilterFunc func(conns *Conn) bool

type Aria interface {
	BroadCast(ctx context.Context, message []byte) error
	BroadCastFilter(ctx context.Context, message []byte, filter FilterFunc) error

	HandleWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	Handle(w http.ResponseWriter, r *http.Request) error

	OnConnect(func(ctx context.Context, conn *Conn) error)
	OnMessage(func(ctx context.Context, conn *Conn, message []byte) error)
	OnMessageBinary(func(ctx context.Context, conn *Conn, message []byte) error)
	OnDisconnect(func(ctx context.Context, conn *Conn) error)
	OnClose(func(ctx context.Context, conn *Conn) error)
	OnError(func(ctx context.Context, conn *Conn, err error))
}

var _ Aria = (*aria)(nil)

// New initialize aria websocket framework instance.
// New can setup option by Functional Option Partten.
// Option func is provided as WithXxx function.
func New(opts ...Option) *aria {
	a := &aria{
		mutex: &sync.Mutex{},
		conns: make(Conns),
	}

	for _, opt := range opts {
		opt.Apply(a)
	}
	return a
}

// Handle is entry-point for accept websocket connection.
// Context is used in websocket communication is initialized in Handle by context.Background().
// If aria accept websocket connection, aria generate goroutine for connection and watch websocket signal.
// When the websocket connection is closed with no problem, the handled connection call OnClose hooked procedure.
// OnClose no happen error, handle return non-error(OnError procedure must not be called).
// When the websocket connection is closed with problem, the handled connection call OnDisconnect hooked procedure.
// OnDisconnect no happen error, handle return non-error(OnError procedure must not be called).
func (a *aria) Handle(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	if err := a.HandleWithContext(ctx, w, r); err != nil {
		return err
	}

	return nil
}

// HandleWithContext is entry-point for accept websocket connection and, can specify context.Context.
func (a *aria) HandleWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	a.conns[ariaConn] = struct{}{}

	if a.onConnect != nil {
		if err := a.onConnect(ctx, ariaConn); err != nil {
			a.onError(ctx, ariaConn, err)
			return err
		}
	}

	go func(c *Conn) {
		if err := a.run(ctx, c); err != nil {
			if a.onError != nil {
				a.onError(ctx, c, err)
			}
		}
	}(ariaConn)

	return nil
}

const cannotClassifyStatus = -1

func (a *aria) run(ctx context.Context, c *Conn) (e error) {
	defer func() {
		a.mutex.Lock()
		a.conns.Remove(c)
		a.mutex.Unlock()
	}()

	for {
		typ, msg, err := c.conn.Read(ctx)
		if err != nil {
			switch websocket.CloseStatus(err) {
			case websocket.StatusNormalClosure, websocket.StatusGoingAway, websocket.StatusNoStatusRcvd:
				if a.onClose != nil {
					if cerr := a.onClose(ctx, c); cerr != nil {
						return cerr
					}

					return nil
				}
			case websocket.StatusAbnormalClosure, websocket.StatusMessageTooBig, websocket.StatusBadGateway, cannotClassifyStatus, websocket.StatusInvalidFramePayloadData:
				if a.onDisconnect != nil {
					if derr := a.onDisconnect(ctx, c); derr != nil {
						return derr
					}
				}
			}

			e = err
			return
		}

		switch typ {
		case websocket.MessageText:
			if a.onMessage != nil {
				if err := a.onMessage(ctx, c, msg); err != nil {
					e = err
					return
				}
			}
		case websocket.MessageBinary:
			if a.onMessageBinary != nil {
				if err := a.onMessageBinary(ctx, c, msg); err != nil {
					e = err
					return
				}
			}
		}
	}
}

// BroadCast send message to connections that is not closed.
// When happend error for sending message, call onClose Procedure.
// BoardCast may happen multiple errors, so BroadCast error is joined by errors.Join.
func (a *aria) BroadCast(ctx context.Context, message []byte) error {
	failedConns := make([]*Conn, 0, len(a.conns))

	var errs []error
	a.mutex.Lock()
	for conn := range a.conns {
		if err := conn.conn.Write(ctx, websocket.MessageText, message); err != nil {
			failedConns = append(failedConns, conn)
			if a.onError != nil {
				a.onError(ctx, conn, err)
			}

			errs = append(errs, err)
		}
	}

	for _, conn := range failedConns {
		a.conns.Remove(conn)
	}
	a.mutex.Unlock()

	return errors.Join(errs...)
}

// BroadCastFilter send message to filtered connections that is not closed.
// A connection FilterFunc result is true, BroadCastFilter send message for the connection.
// BoardCastFilter may happen multiple errors, so BroadCastFilter error is joined by errors.Join too.
func (a *aria) BroadCastFilter(ctx context.Context, message []byte, filter FilterFunc) error {
	failedConns := make([]*Conn, 0, len(a.conns))

	var errs []error
	a.mutex.Lock()
	for conn := range a.conns {
		if filter(conn) {
			if err := conn.conn.Write(ctx, websocket.MessageText, message); err != nil {
				failedConns = append(failedConns, conn)
				if a.onError != nil {
					a.onError(ctx, conn, err)
				}

				errs = append(errs, err)
			}
		}
	}

	for _, conn := range failedConns {
		a.conns.Remove(conn)
	}
	a.mutex.Unlock()

	return errors.Join(errs...)
}

// OnConect setup hooks procedure when websocket connection is established.
func (a *aria) OnConnect(fn func(ctx context.Context, conn *Conn) error) {
	a.onConnect = fn
}

// OnMessage setup hooks procedure when recieved message by established connections.
func (a *aria) OnMessage(fn func(ctx context.Context, conn *Conn, message []byte) error) {
	a.onMessage = fn
}

// OnClose setup hooks procedure when the connection is closed.
// The behavior of OnClose or OnDisconnect is tended to conflict.
// The difference between OnClose and OnDisconnect is mentioned in Handle.
// Please refer to Handle method about the difference.
func (a *aria) OnClose(fn func(ctx context.Context, conn *Conn) error) {
	a.onClose = fn
}

// OnDisconnect setup hooks procedure when the connection is disconnected.
// The behavior of OnClose or OnDisconnect is tended to conflict.
// The difference between OnClose and OnDisconnect is mentioned in Handle.
// Please refer to Handle method about the difference.
func (a *aria) OnDisconnect(fn func(ctx context.Context, conn *Conn) error) {
	a.onDisconnect = fn
}

// OnMessageBinary setup hooks procedure when recieved binary message by established connections.
// Binary message is any bytes. If you send in JavaScript, you will write following code.
// ws = new WebSocket()
// let data = new uint8array(hogehoge)
// ws.send(data)
func (a *aria) OnMessageBinary(fn func(ctx context.Context, conn *Conn, message []byte) error) {
	a.onMessageBinary = fn
}

// OnError setup procedure when occured error in websocket connections.
func (a *aria) OnError(fn func(ctx context.Context, conn *Conn, err error)) {
	a.onError = fn
}
