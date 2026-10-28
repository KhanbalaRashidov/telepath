package permission

import (
	"context"
	"testing"
	"time"
)

func TestBrokerApprove(t *testing.T) {
	b := NewBroker(time.Second)
	go func() {
		// Wait until the request is registered, then approve.
		waitPending(b, "id1")
		b.Resolve("id1", true, "ok")
	}()
	d := b.Await(context.Background(), "id1")
	if !d.Allow || d.Note != "ok" {
		t.Fatalf("got %#v, want allow/ok", d)
	}
}

func TestBrokerDeny(t *testing.T) {
	b := NewBroker(time.Second)
	go func() {
		waitPending(b, "id2")
		b.Resolve("id2", false, "nope")
	}()
	d := b.Await(context.Background(), "id2")
	if d.Allow || d.Note != "nope" {
		t.Fatalf("got %#v, want deny/nope", d)
	}
}

func TestBrokerTimeoutDenies(t *testing.T) {
	b := NewBroker(30 * time.Millisecond)
	start := time.Now()
	d := b.Await(context.Background(), "id3")
	if d.Allow {
		t.Fatal("timeout must deny")
	}
	if time.Since(start) < 25*time.Millisecond {
		t.Fatal("returned before the timeout elapsed")
	}
}

func TestBrokerContextCancelDenies(t *testing.T) {
	b := NewBroker(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		waitPending(b, "id4")
		cancel()
	}()
	d := b.Await(ctx, "id4")
	if d.Allow {
		t.Fatal("cancelled context must deny")
	}
}

func TestResolveUnknownIsNoop(t *testing.T) {
	b := NewBroker(time.Second)
	if b.Resolve("missing", true, "") {
		t.Fatal("resolving an unknown id should report false")
	}
}

func waitPending(b *Broker, id string) {
	for i := 0; i < 200; i++ {
		if b.Pending(id) {
			return
		}
		time.Sleep(time.Millisecond)
	}
}
