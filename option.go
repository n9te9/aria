package aria

import (
	"context"

	"github.com/coder/websocket"
)

type Option interface {
	Apply(*aria)
}

func WithSubprotocols(protocols ...string) Option {
	return &subprotocolsOption{protocols: protocols}
}

type subprotocolsOption struct {
	protocols []string
}

func (o *subprotocolsOption) Apply(a *aria) {
	a.subprotocols = o.protocols
}

func WithInsecureSkipVerify(skip bool) Option {
	return &insecureSkipVerifyOption{skip: skip}
}

type insecureSkipVerifyOption struct {
	skip bool
}

func (o *insecureSkipVerifyOption) Apply(a *aria) {
	a.insecureSkipVerify = o.skip
}

func WithOriginPatterns(patterns ...string) Option {
	return &originPatternsOption{patterns: patterns}
}

type originPatternsOption struct {
	patterns []string
}

func (o *originPatternsOption) Apply(a *aria) {
	a.originPatterns = o.patterns
}

func WithComporessionMode(mode int) Option {
	return &compressionModeOption{mode: mode}
}

type compressionModeOption struct {
	mode int
}

func (o *compressionModeOption) Apply(a *aria) {
	a.compressionMode = websocket.CompressionMode(o.mode)
}

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

func WithOnPingReceived(fn func(ctx context.Context, payload []byte) bool) Option {
	return onPingReceivedOption(fn)
}

type onPongReceivedOption func(ctx context.Context, payload []byte)

func (o onPongReceivedOption) Apply(a *aria) {
	a.onPongReceived = o
}

func WithOnPongReceived(fn func(ctx context.Context, payload []byte)) Option {
	return onPongReceivedOption(fn)
}
