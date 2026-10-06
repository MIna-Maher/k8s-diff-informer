package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestShutdownDuringRetry(t *testing.T) {
	q := NewMemoryQueue(1, 1, nil)
	started := make(chan struct{})
	var once sync.Once
	q.Start(func(ctx context.Context, task *Task) error {
		once.Do(func() { close(started) })
		<-ctx.Done()
		return fmt.Errorf("delivery interrupted")
	})
	if err := q.Enqueue(&Task{Type: "test", MaxRetries: 1}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not start")
	}
	done := make(chan struct{})
	go func() { q.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not finish")
	}
	if err := q.Enqueue(&Task{Type: "test"}); err == nil {
		t.Fatal("accepted task after shutdown")
	}
}
