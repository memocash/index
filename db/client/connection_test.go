package client

import (
	"fmt"
	"sync"
	"testing"

	"google.golang.org/grpc"
)

// dial returns a lazy (non-blocking) client connection; no server needs to
// exist for cache-lifecycle purposes.
func dial(t *testing.T, host string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing %s; %v", host, err)
	}
	return conn
}

// TestConnHandlerReset covers the cache lifecycle: add/get roundtrip, reset
// dropping everything, and the cache being usable again after a reset. The
// handler is a synchronous mutex-protected map — there is no background
// goroutine whose termination could leak across resets, which is the property
// that replaced the old channel/goroutine design.
func TestConnHandlerReset(t *testing.T) {
	var h = new(ConnHandler)
	if got := h.Get("a:1"); got != nil {
		t.Fatalf("expected nil from empty handler, got %v", got)
	}
	h.Reset() // reset of an empty handler must be a safe no-op
	connA := dial(t, "a:1")
	h.Add("a:1", connA)
	if got := h.Get("a:1"); got != connA {
		t.Fatalf("expected cached conn back, got %v", got)
	}
	h.Reset()
	if got := h.Get("a:1"); got != nil {
		t.Fatalf("expected nil after reset, got %v", got)
	}
	connB := dial(t, "b:2")
	h.Add("b:2", connB) // cache must be usable again after reset
	if got := h.Get("b:2"); got != connB {
		t.Fatalf("expected re-added conn after reset, got %v", got)
	}
	h.Reset()
}

// TestConnHandlerConcurrent hammers Add/Get/Reset from many goroutines; run
// under -race this pins the synchronization of the cache (the old design's
// Get raced the handler goroutine's map writes).
func TestConnHandlerConcurrent(t *testing.T) {
	var h = new(ConnHandler)
	var conns [4]*grpc.ClientConn
	for i := range conns {
		conns[i] = dial(t, fmt.Sprintf("concurrent:%d", i))
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				host := fmt.Sprintf("concurrent:%d", i%len(conns))
				switch (worker + i) % 3 {
				case 0:
					h.Add(host, conns[i%len(conns)])
				case 1:
					h.Get(host)
				default:
					h.Reset()
				}
			}
		}(worker)
	}
	wg.Wait()
	h.Reset()
}

// TestResetConnectionsPackageLevel exercises the package-level reset used by
// the test suite between same-process suite runs, including when nothing has
// ever been cached.
func TestResetConnectionsPackageLevel(t *testing.T) {
	ResetConnections() // must be safe with an untouched cache
	connHandler.Add("pkg:1", dial(t, "pkg:1"))
	ResetConnections()
	if got := connHandler.Get("pkg:1"); got != nil {
		t.Fatalf("expected nil after package-level reset, got %v", got)
	}
}
