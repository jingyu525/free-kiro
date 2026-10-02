package visualize

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// watchTickInterval is the polling cadence for watchChanges. Exported
// only as a package-level const so tests can shorten it (or override
// the ticker in a refactor). Kept short enough that authors see their
// edits within a second; long enough to debounce filesystem noise.
const watchTickInterval = 2 * time.Second

// subscribe registers a new SSE client channel on this Server. The
// returned channel is buffered so a slow client never blocks broadcast.
//
// Acquires s.notifierMu to serialise against concurrent
// unsubscribe / broadcast calls (S1: fix the package-level race).
func (s *Server) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	s.notifierMu.Lock()
	s.subscribers = append(s.subscribers, ch)
	s.notifierMu.Unlock()
	return ch
}

// unsubscribe removes a client channel. Idempotent — a second call
// with the same channel is a no-op (matches the pre-fix contract).
//
// We deliberately do NOT close(ch): the SSE handler exits via its
// request context, and a concurrent broadcast() may still be holding
// a snapshot reference to the channel. Closing under that race would
// panic with "send on closed channel" inside broadcast. The channel
// is garbage-collected once the handler returns.
func (s *Server) unsubscribe(ch chan struct{}) {
	s.notifierMu.Lock()
	defer s.notifierMu.Unlock()
	for i, c := range s.subscribers {
		if c == ch {
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			return
		}
	}
}

// broadcast signals every subscriber that the dashboard should refresh.
// Takes a snapshot of the subscriber list under the mutex so we don't
// hold the lock while sending to (potentially slow) client channels.
func (s *Server) broadcast() {
	s.notifierMu.Lock()
	subs := make([]chan struct{}, len(s.subscribers))
	copy(subs, s.subscribers)
	s.notifierMu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
			// Subscriber buffer full — drop the signal; the client will
			// pick up the next refresh on its 5s polling fallback.
		}
	}
}

// watchChanges polls .kiro/ for file modifications and broadcasts a
// refresh signal when something changes. Exits when s.stop is closed
// and closes s.watcherRunning so Shutdown() can wait for it.
func (s *Server) watchChanges() {
	s.watcherStartCount.Add(1)
	defer close(s.watcherRunning)
	last := s.collectMtimes()
	tick := time.NewTicker(watchTickInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-tick.C:
			cur := s.collectMtimes()
			if !mtimesEqual(last, cur) {
				last = cur
				s.broadcast()
			}
		}
	}
}

// collectMtimes returns a path → mtime map for every file under .kiro/.
// Missing directory → empty map (treat as "no changes").
func (s *Server) collectMtimes() map[string]time.Time {
	out := map[string]time.Time{}
	root := s.ws.KiroDir()
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			out[path] = info.ModTime()
		}
		return nil
	})
	return out
}

func mtimesEqual(a, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || !va.Equal(vb) {
			return false
		}
	}
	return true
}

// handleEvents implements the SSE endpoint. Clients receive an initial
// ping plus a refresh signal every time watchChanges fires.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	// Cache-Control is owned by withCacheHeaders middleware (sets
	// "no-cache, no-store, must-revalidate" for /api/events so proxies
	// don't buffer SSE). Don't re-set it here — that would clobber
	// the middleware's stricter no-store directive.

	// Best-effort: drop the client when it disconnects.
	ch := s.subscribe()
	defer s.unsubscribe(ch)

	// Initial ping so EventSource knows the stream is live.
	_, _ = fmt.Fprint(w, "event: ping\ndata: ok\n\n")
	flusher.Flush()

	// Start the watcher exactly once per Server lifetime. Subsequent
	// SSE clients reuse the running watcher (S8: don't spawn duplicates).
	s.watcherOnce.Do(func() {
		go s.watchChanges()
	})

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			if _, err := fmt.Fprint(w, "event: refresh\ndata: 1\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// keep imports referenced after split (os is used by collectMtimes, etc.).
var _ = context.Background
