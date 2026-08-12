package client

import (
	"google.golang.org/grpc"
	"sync"
)

// ConnHandler caches one client connection per host. A mutex-protected map
// rather than a goroutine/channel: adds, lookups, and resets are synchronous,
// so there is no background goroutine to leak (or race with) across resets.
type ConnHandler struct {
	mu          sync.Mutex
	connections map[string]*grpc.ClientConn
}

func (h *ConnHandler) Get(host string) *grpc.ClientConn {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.connections[host]
}

func (h *ConnHandler) Add(host string, conn *grpc.ClientConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connections == nil {
		h.connections = make(map[string]*grpc.ClientConn)
	}
	h.connections[host] = conn
}

// Reset synchronously drops every cached connection, closing each one, so
// the next Get misses and callers dial fresh. Idempotent.
func (h *ConnHandler) Reset() {
	h.mu.Lock()
	var old = h.connections
	h.connections = nil
	h.mu.Unlock()
	for _, conn := range old {
		_ = conn.Close()
	}
}

// ResetConnections drops every cached client connection so the next call
// dials fresh. For tests that stop and restart servers on the same ports
// within one process: a cached connection to a stopped server can hand its
// first caller an Unavailable/EOF before gRPC notices and reconnects.
func ResetConnections() {
	connHandler.Reset()
}
