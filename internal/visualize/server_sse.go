package visualize

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// notifier fans out change events to all subscribed SSE clients. A
// single broadcast wakes every watcher, so they re-render the dashboard.
var notifier = struct {
	subscribers []chan struct{}
}{}

// subscribe registers a new SSE client channel.
func subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	notifier.subscribers = append(notifier.subscribers, ch)
	return ch
}

// unsubscribe removes a client channel. Idempotent.
func unsubscribe(ch chan struct{}) {
	for i, c := range notifier.subscribers {
		if c == ch {
			notifier.subscribers = append(notifier.subscribers[:i], notifier.subscribers[i+1:]...)
			close(c)
			return
		}
	}
}

// broadcast signals every subscriber that the dashboard should refresh.
func broadcast() {
	for _, ch := range notifier.subscribers {
		select {
		case ch <- struct{}{}:
		default:
			// Subscriber buffer full — drop the signal; the client will
			// pick up the next refresh on its 5s polling fallback.
		}
	}
}

// watchChanges polls .kiro/ for file modifications and broadcasts a
// refresh signal when something changes. Exits when s.stop is closed.
func (s *Server) watchChanges() {
	defer close(s.done)
	last := s.collectMtimes()
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-tick.C:
			cur := s.collectMtimes()
			if !mtimesEqual(last, cur) {
				last = cur
				broadcast()
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
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Best-effort: drop the client when it disconnects.
	ch := subscribe()
	defer unsubscribe(ch)

	// Initial ping so EventSource knows the stream is live.
	fmt.Fprint(w, "event: ping\ndata: ok\n\n")
	flusher.Flush()

	// Start the watcher if not already running.
	go s.watchChanges()

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

// ensure strings stays referenced after splitting files.
var _ = strings.Contains
