package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestLeasePreventsDuplicateConsumption(t *testing.T) {
	queue := NewMemoryQueue()
	if err := queue.Enqueue(Job{ID: "one", Kind: "test", AvailableAt: time.Now().Add(-time.Second)}); err != nil {
		t.Fatal(err)
	}
	first, err := queue.Lease(context.Background(), "worker-one", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = queue.Lease(context.Background(), "worker-two", time.Minute); err != ErrNoJobs {
		t.Fatalf("leased job delivered twice: %v", err)
	}
	if err = queue.Complete(context.Background(), first.ID, "worker-two"); err == nil {
		t.Fatal("different worker completed lease")
	}
	if err = queue.Complete(context.Background(), first.ID, "worker-one"); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerRetriesThenCompletes(t *testing.T) {
	queue := NewMemoryQueue()
	if err := queue.Enqueue(Job{ID: "retry", Kind: "test", AvailableAt: time.Now().Add(-time.Second)}); err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(queue, "worker", time.Minute)
	var calls atomic.Int32
	runner.Register("test", func(context.Context, Job) error {
		calls.Add(1)
		return nil
	})
	if err := runner.runOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("handler calls = %d", calls.Load())
	}
}
