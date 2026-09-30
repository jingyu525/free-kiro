package visualize

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// stubWS satisfies the WorkspacePaths interface with the minimum
// surface Server tests need. Methods return zero values because the
// SSE / shutdown paths don't actually consult the workspace.
type stubWS struct{}

func (stubWS) KiroDir() string       { return "" }
func (stubWS) SpecDir(string) string { return "" }
func (stubWS) Root() string          { return "" }
func (stubWS) ReadCurrent() string   { return "" }

// TestServer_ShutdownWithoutWatcher covers AC-3 (V1): Shutdown must
// return within shutdownTimeout when watchChanges was never started.
// Pre-fix the call hung forever on `<-s.done`.
func TestServer_ShutdownWithoutWatcher(t *testing.T) {
	s := NewServer(":0", stubWS{}, nil)

	start := time.Now()
	done := make(chan error, 1)
	go func() { done <- s.Shutdown() }()
	select {
	case err := <-done:
		if elapsed := time.Since(start); elapsed > shutdownTimeout+time.Second {
			t.Errorf("Shutdown took %v; expected <= %v (deadline)", elapsed, shutdownTimeout+time.Second)
		}
		// Shutdown on a never-listened server returns http.ErrServerClosed,
		// which we accept as a clean shutdown.
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Shutdown error: %v", err)
		}
	case <-time.After(shutdownTimeout + 2*time.Second):
		t.Fatal("Shutdown did not return within budget — V1 regression")
	}
}

// TestServer_ConcurrentSubscribeBroadcast covers AC-1 (S1): under
// concurrent subscribe / unsubscribe / broadcast, the package-level
// notifier previously triggered -race findings. Run with
// `go test -race` to verify; the test itself just exercises the
// workload (50 goroutines × 200 iterations each).
func TestServer_ConcurrentSubscribeBroadcast(t *testing.T) {
	s := NewServer(":0", stubWS{}, nil)

	const goroutines = 50
	const iterations = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				ch := s.subscribe()
				s.broadcast()
				s.unsubscribe(ch)
			}
		}()
	}
	wg.Wait()

	// Sanity: all subscribers should have unsubscribed themselves.
	s.notifierMu.Lock()
	remaining := len(s.subscribers)
	s.notifierMu.Unlock()
	if remaining != 0 {
		t.Errorf("expected 0 subscribers after goroutines finish; got %d", remaining)
	}
}

// TestServer_WatcherStartedOnce covers AC-2 (S8): even when multiple
// SSE clients trigger handleEvents, watchChanges should be spawned
// exactly once per Server lifetime. We invoke handleEvents directly
// against multiple synthetic HTTP requests and read watcherStartCount.
func TestServer_WatcherStartedOnce(t *testing.T) {
	s := NewServer(":0", stubWS{}, nil)

	for range 3 {
		req := httptest.NewRequest("GET", "/api/events", nil)
		ctx, cancel := context.WithCancel(context.Background())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.handleEvents(w, req)
		}()
		// Brief delay so the watcherOnce.Do fires before we cancel.
		time.Sleep(20 * time.Millisecond)
		cancel()
		<-done
	}

	// Settle: wait a hair to let any erroneously-spawned watchers
	// start. Then assert exactly 1.
	time.Sleep(50 * time.Millisecond)
	if got := s.watcherStartCount.Load(); got != 1 {
		t.Errorf("expected watcherStartCount == 1 after 3 handleEvents calls; got %d", got)
	}

	// Clean shutdown so the spawned watcher exits cleanly.
	_ = s.Shutdown()
}

// TestServer_URL covers AC-3: URL() must produce a syntactically valid
// http URL across IPv4, IPv6, and port-only listen addrs. The
// pre-fix implementation concatenated "http://" + the port substring
// of the listen address, yielding "http://8080" for any input.
func TestServer_URL(t *testing.T) {
	cases := []struct {
		name string
		addr string
		want string
	}{
		{"ipv4 port", "127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"ipv6 literal", "[::1]:8080", "http://[::1]:8080"},
		{"ipv6 unspecified", "[::]:8080", "http://[::]:8080"},
		{"localhost port", "localhost:9090", "http://localhost:9090"},
		// Port-only input (no host) — fall back to passing Addr through
		// net.SplitHostPort's failure path.
		{"port only", "8080", "http://8080"},
		{"hostname no port", "example.com", "http://example.com"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewServer(c.addr, stubWS{}, nil)
			if got := s.URL(); got != c.want {
				t.Errorf("URL() = %q; want %q", got, c.want)
			}
		})
	}
}
