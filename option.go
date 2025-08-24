package aria

import (
	"context"

	"github.com/coder/websocket"
)

// Option represents a configuration option for an aria instance.
// Options can be passed to New() to customize server behavior.
type Option interface {
	Apply(*aria)
}

// WithSubprotocols specifies the WebSocket subprotocols
// that the server is willing to accept during the handshake.
func WithSubprotocols(protocols ...string) Option {
	return &subprotocolsOption{protocols: protocols}
}

type subprotocolsOption struct {
	protocols []string
}

func (o *subprotocolsOption) Apply(a *aria) {
	a.subprotocols = o.protocols
}

// WithInsecureSkipVerify configures whether to skip TLS certificate
// verification during the WebSocket handshake.
// This is useful for development or when using self-signed certificates.
func WithInsecureSkipVerify(skip bool) Option {
	return &insecureSkipVerifyOption{skip: skip}
}

type insecureSkipVerifyOption struct {
	skip bool
}

func (o *insecureSkipVerifyOption) Apply(a *aria) {
	a.insecureSkipVerify = o.skip
}

// WithOriginPatterns sets the allowed origin patterns for validating
// the WebSocket client's Origin header.
func WithOriginPatterns(patterns ...string) Option {
	return &originPatternsOption{patterns: patterns}
}

type originPatternsOption struct {
	patterns []string
}

func (o *originPatternsOption) Apply(a *aria) {
	a.originPatterns = o.patterns
}

// WithComporessionMode sets the compression mode for WebSocket messages.
// Use values from websocket.CompressionMode.
func WithComporessionMode(mode int) Option {
	return &compressionModeOption{mode: mode}
}

type compressionModeOption struct {
	mode int
}

func (o *compressionModeOption) Apply(a *aria) {
	a.compressionMode = websocket.CompressionMode(o.mode)
}

// WithCompressionThreshold sets the minimum message size (in bytes)
// required before compression is applied.
func WithCompressionThreshold(threshold int) Option {
	return compressionThresholdOption(threshold)
}

type compressionThresholdOption int

func (o compressionThresholdOption) Apply(a *aria) {
	a.compressionThreshold = int(o)
}

type onPingReceivedOption func(ctx context.Context, payload []byte) bool

func (o onPingReceivedOption) Apply(a *aria) {
	a.onPingReceived = o
}

// WithOnPingReceived registers a handler for incoming Ping frames.
// The callback return value determines whether the library should
// automatically respond with a Pong frame (true = auto respond).
func WithOnPingReceived(fn func(ctx context.Context, payload []byte) bool) Option {
	return onPingReceivedOption(fn)
}

type onPongReceivedOption func(ctx context.Context, payload []byte)

func (o onPongReceivedOption) Apply(a *aria) {
	a.onPongReceived = o
}

// WithOnPongReceived registers a handler for incoming Pong frames.
func WithOnPongReceived(fn func(ctx context.Context, payload []byte)) Option {
	return onPongReceivedOption(fn)
}
