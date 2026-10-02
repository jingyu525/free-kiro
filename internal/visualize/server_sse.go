package visualize

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchTickInterval is the polling cadence used only when fsnotify is
// unavailable (containers / network filesystems). With a working inotify
// watch we get < 200ms notification latency via OS-level events instead.
//
// Exported as a package-level const so tests can shorten it.
const watchTickInterval = 2 * time.Second

// watchFsnotifyDebounce is the quiet period after the last fs event
// before we broadcast. Coalesces the editor's truncate+write two-step.
const watchFsnotifyDebounce = 300 * time.Millisecond

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
// Also updates lastRefreshAt so /api/health can surface it.
func (s *Server) broadcast() {
	s.notifierMu.Lock()
	subs := make([]chan struct{}, len(s.subscribers))
	copy(subs, s.subscribers)
	s.notifierMu.Unlock()
	s.lastRefreshAt.Store(time.Now().UTC().UnixNano())
	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
			// Subscriber buffer full — drop the signal; the client will
			// pick up the next refresh on its 5s polling fallback.
		}
	}
}

// subscribersCount returns the number of currently-connected SSE
// clients. Used by /api/health. Safe to call concurrently.
func (s *Server) subscribersCount() int {
	s.notifierMu.Lock()
	defer s.notifierMu.Unlock()
	return len(s.subscribers)
}

// watchChanges is the single per-Server watcher loop started by
// watcherOnce. It tries fsnotify first (OS-level events, < 200ms
// latency); if fsnotify.NewWatcher fails (no inotify/FSEvents — e.g.
// in some container/network-FS mounts), it falls back to a 2s mtime
// polling ticker so the dashboard still gets refresh signals.
//
// Exits when s.stop is closed and closes s.watcherRunning so Shutdown()
// can wait for it.
func (s *Server) watchChanges() {
	s.watcherStartCount.Add(1)
	defer close(s.watcherRunning)

	root := s.ws.KiroDir()
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		s.watchChangesFallback(root)
		return
	}
	defer func() { _ = fs.Close() }()

	if err := addKiroTree(fs, root); err != nil {
		// Couldn't add the root; fall back to polling so the
		// dashboard still gets *some* signal (slower but functional).
		s.watchChangesFallback(root)
		return
	}

	var (
		timer   *time.Timer
		timerCh <-chan time.Time
	)
	for {
		select {
		case <-s.stop:
			if timer != nil {
				timer.Stop()
			}
			return
		case ev, ok := <-fs.Events:
			if !ok {
				return
			}
			if !isMeaningfulOp(ev.Op) {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(watchFsnotifyDebounce)
			timerCh = timer.C
		case <-timerCh:
			timerCh = nil
			timer = nil
			s.broadcast()
		case err, ok := <-fs.Errors:
			if !ok {
				return
			}
			// fsnotify errors are usually transient (inode gone, file
			// removed). Log to stderr and keep the loop alive so the
			// next event still triggers a refresh.
			fmt.Fprintf(os.Stderr, "[watch] fsnotify: %v\n", err)
		}
	}
}

// watchChangesFallback polls .kiro/ for mtime changes every
// watchTickInterval. Used when fsnotify.NewWatcher fails or the root
// can't be added to the watcher (containers / network FS / read-only
// mounts).
func (s *Server) watchChangesFallback(root string) {
	last := collectMtimes(root)
	tick := time.NewTicker(watchTickInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-tick.C:
			cur := collectMtimes(root)
			if !mtimesEqual(last, cur) {
				last = cur
				s.broadcast()
			}
		}
	}
}

// isMeaningfulOp returns true for fsnotify events that signal an
// actual content change worth broadcasting.
func isMeaningfulOp(op fsnotify.Op) bool {
	switch op {
	case fsnotify.Write, fsnotify.Create, fsnotify.Remove, fsnotify.Rename:
		return true
	}
	return false
}

// addKiroTree adds root + every subdirectory under it to the watcher.
// Skips noisy editor / VCS directories. fsnotify doesn't recurse on
// its own — we have to enumerate.
func addKiroTree(fs *fsnotify.Watcher, root string) error {
	if err := fs.Add(root); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Permission errors on individual subdirs are non-fatal;
			// skip and continue.
			return nil
		}
		if !d.IsDir() || path == root {
			return nil
		}
		name := d.Name()
		if name == ".git" || name == "node_modules" {
			return filepath.SkipDir
		}
		if err := fs.Add(path); err != nil {
			// Subdir watch failure shouldn't kill the whole watcher.
			return nil
		}
		return nil
	})
}

// collectMtimes returns a path → mtime map for every file under root.
// Missing directory → empty map (treat as "no changes").
func collectMtimes(root string) map[string]time.Time {
	out := map[string]time.Time{}
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