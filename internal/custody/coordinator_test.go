package custody

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"material-lab-platform/internal/domain"
)

func TestCoordinatorAllowsOneConcurrentTerminalTransition(t *testing.T) {
	transfer, err := domain.NewCustodyTransfer("sub", "sender", "receiver", "a", "b", domain.CustodyHandOver, "sealed", "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	coordinator := &Coordinator{}
	start := make(chan struct{})
	var successes atomic.Int32
	var wait sync.WaitGroup
	for i := 0; i < 16; i++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			if index%2 == 0 {
				if coordinator.Confirm(&transfer, "receiver", 1, time.Now()) == nil {
					successes.Add(1)
				}
			} else if _, transitionErr := coordinator.Reverse(&transfer, "manager", "correction", 1, time.Now()); transitionErr == nil {
				successes.Add(1)
			}
		}(i)
	}
	close(start)
	wait.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful terminal transitions = %d", successes.Load())
	}
	if transfer.Revision != 2 {
		t.Fatalf("revision = %d", transfer.Revision)
	}
}
