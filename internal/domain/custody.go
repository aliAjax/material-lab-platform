package domain

import (
	"fmt"
	"strings"
	"time"
)

type CustodyAction string

const (
	CustodyPickup   CustodyAction = "pickup"
	CustodyHandOver CustodyAction = "transfer"
	CustodyReturn   CustodyAction = "return"
	CustodySeal     CustodyAction = "seal"
	CustodyReversal CustodyAction = "reversal"
)

type CustodyStatus string

const (
	CustodyPending   CustodyStatus = "pending"
	CustodyConfirmed CustodyStatus = "confirmed"
	CustodyReversed  CustodyStatus = "reversed"
)

type CustodyTransfer struct {
	ID             string        `json:"id"`
	SubsampleID    string        `json:"subsampleId"`
	Action         CustodyAction `json:"action"`
	FromUserID     string        `json:"fromUserId"`
	ToUserID       string        `json:"toUserId"`
	FromLocationID string        `json:"fromLocationId"`
	ToLocationID   string        `json:"toLocationId"`
	Condition      string        `json:"condition"`
	Notes          string        `json:"notes,omitempty"`
	Status         CustodyStatus `json:"status"`
	OriginalID     string        `json:"originalId,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
	ConfirmedAt    time.Time     `json:"confirmedAt,omitempty"`
	Revision       uint64        `json:"revision"`
}

func NewCustodyTransfer(subsample, from, to, fromLocation, toLocation string, action CustodyAction, condition, notes string, now time.Time) (CustodyTransfer, error) {
	if from == to {
		return CustodyTransfer{}, fmt.Errorf("%w: both parties must differ", ErrValidation)
	}
	if strings.TrimSpace(condition) == "" {
		return CustodyTransfer{}, fmt.Errorf("%w: condition required", ErrValidation)
	}
	switch action {
	case CustodyPickup, CustodyHandOver, CustodyReturn, CustodySeal:
	default:
		return CustodyTransfer{}, fmt.Errorf("%w: unsupported custody action", ErrValidation)
	}
	return CustodyTransfer{ID: NewID(), SubsampleID: subsample, FromUserID: from, ToUserID: to, FromLocationID: fromLocation, ToLocationID: toLocation, Action: action, Condition: condition, Notes: notes, Status: CustodyPending, CreatedAt: now, Revision: 1}, nil
}

func (t *CustodyTransfer) Confirm(actor string, now time.Time) error {
	return t.ConfirmRevision(actor, now, t.Revision)
}

func (t *CustodyTransfer) ConfirmRevision(actor string, now time.Time, expected uint64) error {
	if t == nil {
		return ErrNotFound
	}
	if expected == 0 {
		return fmt.Errorf("%w: expected revision required", ErrValidation)
	}
	if t.Status != CustodyPending {
		return ErrConflict
	}
	if t.Revision != expected {
		return ErrConflict
	}
	if actor != t.ToUserID {
		return ErrForbidden
	}
	t.Status, t.ConfirmedAt = CustodyConfirmed, now
	t.Revision = expected + 1
	return nil
}

func NewReversal(original CustodyTransfer, actor, reason string, now time.Time) (CustodyTransfer, error) {
	if original.Status != CustodyConfirmed {
		return CustodyTransfer{}, ErrInvalidTransition
	}
	if strings.TrimSpace(reason) == "" {
		return CustodyTransfer{}, fmt.Errorf("%w: reason required", ErrValidation)
	}
	return CustodyTransfer{ID: NewID(), SubsampleID: original.SubsampleID, Action: CustodyReversal, FromUserID: actor, ToUserID: original.FromUserID, FromLocationID: original.ToLocationID, ToLocationID: original.FromLocationID, Condition: original.Condition, Notes: reason, Status: CustodyPending, OriginalID: original.ID, CreatedAt: now, Revision: 1}, nil
}
