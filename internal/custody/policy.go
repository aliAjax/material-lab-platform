package custody

import (
	"time"

	"material-lab-platform/internal/domain"
)

type Coordinator struct{}

func (c *Coordinator) Confirm(transfer *domain.CustodyTransfer, actor string, expected uint64, now time.Time) error {
	if transfer == nil {
		return domain.ErrNotFound
	}
	if !CanConfirm(*transfer, actor) {
		return domain.ErrForbidden
	}
	return transfer.ConfirmRevision(actor, now, expected)
}

func (c *Coordinator) Reverse(transfer *domain.CustodyTransfer, actor, reason string, expected uint64, now time.Time) (domain.CustodyTransfer, error) {
	if transfer == nil {
		return domain.CustodyTransfer{}, domain.ErrNotFound
	}
	if !CanReverse(*transfer) {
		return domain.CustodyTransfer{}, domain.ErrInvalidTransition
	}
	reversal, err := domain.NewReversal(*transfer, actor, reason, now)
	if err != nil {
		return domain.CustodyTransfer{}, err
	}
	transfer.Status = domain.CustodyReversed
	transfer.Revision++
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
