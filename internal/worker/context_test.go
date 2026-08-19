package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunnerPropagatesCancellationToActiveHandler(t *testing.T) {
	queue := NewMemoryQueue()
	if err := queue.Enqueue(Job{ID: "blocking", Kind: "blocking", AvailableAt: time.Now().Add(-time.Second)}); err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(queue, "worker", time.Minute)
	entered := make(chan struct{})
	runner.Register("blocking", func(ctx context.Context, _ Job) error { close(entered); <-ctx.Done(); return ctx.Err() })
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("runner error = %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("runner did not propagate cancellation")
	}
}

func TestQueueRejectsCanceledStateTransitions(t *testing.T) {
	queue := NewMemoryQueue()
	if err := queue.Enqueue(Job{ID: "one", Kind: "test", AvailableAt: time.Now().Add(-time.Second)}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := queue.Lease(ctx, "worker", time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("lease error = %v", err)
	}
}
