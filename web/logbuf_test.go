package web

import (
	"fmt"
	"testing"
	"time"
)

func TestLogRingKeepsLastLines(t *testing.T) {
	r := NewLogRing(3)
	for i := 0; i < 5; i++ {
		r.Append(LogLine{Level: "info", Message: fmt.Sprintf("line %d", i)})
	}

	got := r.Backlog()
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	if got[0].Message != "line 2" || got[2].Message != "line 4" {
		t.Errorf("expected lines 2..4, got %q..%q", got[0].Message, got[2].Message)
	}
}

func TestLogRingSubscriberReceives(t *testing.T) {
	r := NewLogRing(10)
	ch, cancel := r.Subscribe(4)
	defer cancel()

	r.Append(LogLine{Level: "info", Message: "hello"})

	select {
	case l := <-ch:
		if l.Message != "hello" {
			t.Errorf("expected 'hello', got %q", l.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive the line")
	}
}

func TestLogRingSlowSubscriberDoesNotBlock(t *testing.T) {
	r := NewLogRing(10)
	_, cancel := r.Subscribe(1) // canale mai letto
	defer cancel()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			r.Append(LogLine{Level: "info", Message: "flood"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Append blocked on a slow subscriber")
	}
}

func TestLogRingCancelRemovesSubscriber(t *testing.T) {
	r := NewLogRing(10)
	_, cancel := r.Subscribe(1)
	cancel()

	if n := r.subscriberCount(); n != 0 {
		t.Errorf("expected 0 subscribers after cancel, got %d", n)
	}
	cancel() // must be idempotent
}
