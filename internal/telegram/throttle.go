package telegram

import (
	"sync"
	"time"
)

// throttle coalesces frequent status updates into at most one flush per
// interval, so the single live status message is edited within Telegram's rate
// limits (never one edit per token). Important events call Flush to push the
// latest text immediately.
type throttle struct {
	interval time.Duration
	flushFn  func(text string)

	mu     sync.Mutex
	latest string
	dirty  bool
	closed bool

	done chan struct{}
}

// newThrottle starts the background flusher. flushFn must be safe to call from a
// goroutine.
func newThrottle(interval time.Duration, flushFn func(string)) *throttle {
	t := &throttle{
		interval: interval,
		flushFn:  flushFn,
		done:     make(chan struct{}),
	}
	go t.loop()
	return t
}

func (t *throttle) loop() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.flushIfDirty()
		case <-t.done:
			return
		}
	}
}

// Update records the latest status text without necessarily flushing it.
func (t *throttle) Update(text string) {
	t.mu.Lock()
	if text != t.latest {
		t.latest = text
		t.dirty = true
	}
	t.mu.Unlock()
}

// Flush pushes the latest text immediately if it changed since the last flush.
func (t *throttle) Flush() {
	t.flushIfDirty()
}

func (t *throttle) flushIfDirty() {
	t.mu.Lock()
	if !t.dirty {
		t.mu.Unlock()
		return
	}
	text := t.latest
	t.dirty = false
	t.mu.Unlock()
	t.flushFn(text)
}

// Stop halts the background flusher and performs a final flush.
func (t *throttle) Stop() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	t.mu.Unlock()

	close(t.done)
	t.flushIfDirty()
}
