package main

import (
	"context"
	"testing"
	"time"
)

func TestWorkerContextPreservesParentCancellation(t *testing.T) {
	parent, stop := context.WithCancel(context.Background())
	ctx, cancel := workerContext(parent)
	defer cancel()
	stop()
	select {
	case <-ctx.Done():
	case <-time.After(250 * time.Millisecond):
		t.Fatal("worker context ignored parent cancellation")
	}
}
