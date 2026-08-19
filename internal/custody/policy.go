package custody

import "material-lab-platform/internal/domain"

func CanConfirm(transfer domain.CustodyTransfer, actor string) bool {
	return transfer.Status == domain.CustodyPending && transfer.ToUserID == actor && transfer.FromUserID != actor
}

func IsCompleted(transfer domain.CustodyTransfer) bool {
	return transfer.Status == domain.CustodyConfirmed
}

func CanReverse(transfer domain.CustodyTransfer) bool {
	return transfer.Status == domain.CustodyConfirmed && transfer.Action != domain.CustodyReversal
}
