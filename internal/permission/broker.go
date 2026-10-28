package permission

import (
	"context"
	"sync"
	"time"
)

// Decision is the resolution of a pending approval.
type Decision struct {
	Allow bool
	Note  string
}

// Broker tracks pending approval requests by id and resolves them when the user
// taps a button (Resolve), the task context is cancelled, or the timeout
// elapses (which denies, so a silent user never auto-approves anything).
//
// It is safe for concurrent use across multiple chats/tasks.
type Broker struct {
	timeout time.Duration

	mu      sync.Mutex
	pending map[string]chan Decision
}

// NewBroker creates a broker with the given approval timeout. A non-positive
// timeout disables the timeout (the request then waits only for Resolve or ctx).
func NewBroker(timeout time.Duration) *Broker {
	return &Broker{
		timeout: timeout,
		pending: make(map[string]chan Decision),
	}
}

// Await registers id as pending and blocks until it is resolved, the context is
// cancelled, or the timeout elapses. Cancellation and timeout both deny.
func (b *Broker) Await(ctx context.Context, id string) Decision {
	ch := make(chan Decision, 1)

	b.mu.Lock()
	b.pending[id] = ch
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.pending, id)
		b.mu.Unlock()
	}()

	var timeout <-chan time.Time
	if b.timeout > 0 {
		t := time.NewTimer(b.timeout)
		defer t.Stop()
		timeout = t.C
	}

	select {
	case d := <-ch:
		return d
	case <-timeout:
		return Decision{Allow: false, Note: "approval timed out"}
	case <-ctx.Done():
		return Decision{Allow: false, Note: "task cancelled"}
	}
}

// Resolve delivers a decision to the pending request id. It reports whether a
// matching pending request existed (false means it already resolved/expired).
func (b *Broker) Resolve(id string, allow bool, note string) bool {
	b.mu.Lock()
	ch, ok := b.pending[id]
	b.mu.Unlock()
	if !ok {
		return false
	}
	select {
	case ch <- Decision{Allow: allow, Note: note}:
		return true
	default:
		// Already resolved; the buffered slot is taken.
		return false
	}
}

// Pending reports whether id is currently awaiting a decision.
func (b *Broker) Pending(id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.pending[id]
	return ok
}
