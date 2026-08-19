package custody

import (
	"fmt"
	"sync"
	"time"

	"material-lab-platform/internal/domain"
)

type Coordinator struct {
	mu sync.Mutex
}

func (c *Coordinator) Confirm(transfer *domain.CustodyTransfer, actor string, expected uint64, now time.Time) error {
	if transfer == nil {
		return domain.ErrNotFound
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !CanConfirm(*transfer, actor) {
		return domain.ErrForbidden
	}
	return transfer.ConfirmRevision(actor, now, expected)
}

func (c *Coordinator) Reverse(transfer *domain.CustodyTransfer, actor, reason string, expected uint64, now time.Time) (domain.CustodyTransfer, error) {
	if transfer == nil {
		return domain.CustodyTransfer{}, domain.ErrNotFound
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !CanReverse(*transfer) {
		return domain.CustodyTransfer{}, domain.ErrInvalidTransition
	}
	if expected == 0 {
		return domain.CustodyTransfer{}, fmt.Errorf("%w: expected revision required", domain.ErrValidation)
	}
	if transfer.Revision != expected {
		return domain.CustodyTransfer{}, domain.ErrConflict
	}
	reversal, err := domain.NewReversal(*transfer, actor, reason, now)
	if err != nil {
		return domain.CustodyTransfer{}, err
	}
	transfer.Status = domain.CustodyReversed
	transfer.Revision = expected + 1
	return reversal, nil
}

func CanConfirm(transfer domain.CustodyTransfer, actor string) bool {
	return transfer.Status == domain.CustodyPending && transfer.ToUserID == actor && transfer.FromUserID != actor
}

func IsCompleted(transfer domain.CustodyTransfer) bool {
	return transfer.Status == domain.CustodyConfirmed
}

func CanReverse(transfer domain.CustodyTransfer) bool {
	return transfer.Status == domain.CustodyConfirmed && transfer.Action != domain.CustodyReversal
}
