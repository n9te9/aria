package aria

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConn_SetGetDelete(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		value  any
		expect any
	}{
		{
			name:   "store and retrieve string",
			key:    "foo",
			value:  "bar",
			expect: "bar",
		},
		{
			name:   "store and retrieve int",
			key:    "num",
			value:  123,
			expect: 123,
		},
		{
			name:   "store and retrieve struct",
			key:    "obj",
			value:  struct{ X int }{X: 42},
			expect: struct{ X int }{X: 42},
		},
	}

	// use nil websocket.Conn because Set/Get/Delete don't touch it
	c := &Conn{store: sync.Map{}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Set(tt.key, tt.value)
			v, ok := c.Get(tt.key)
			assert.True(t, ok)
			assert.Equal(t, tt.expect, v)

			c.Delete(tt.key)
			_, ok = c.Get(tt.key)
			assert.False(t, ok)
		})
	}
}

func TestConns_RemoveAndSlice(t *testing.T) {
	c1 := &Conn{store: sync.Map{}}
	c2 := &Conn{store: sync.Map{}}

	conns := make(Conns)
	conns[c1] = struct{}{}
	conns[c2] = struct{}{}

	tests := []struct {
		name         string
		removeTarget *Conn
		expectLen    int
	}{
		{
			name:         "remove existing conn",
			removeTarget: c1,
			expectLen:    1,
		},
		{
			name:         "remove non-existing conn",
			removeTarget: &Conn{store: sync.Map{}},
			expectLen:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conns.Remove(tt.removeTarget)
			assert.Equal(t, tt.expectLen, len(conns))
			slice := conns.Slice()
			assert.Equal(t, tt.expectLen, len(slice))
		})
	}
}
