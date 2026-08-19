package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRevisionGuardProtectsCustodyMutation(t *testing.T) {
	now := time.Now().UTC()
	var missing *CustodyTransfer
	if err := missing.ConfirmRevision("receiver", now, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("nil transfer error = %v, want not found", err)
	}
	transfer, err := NewCustodyTransfer("sub", "sender", "receiver", "a", "b", CustodyHandOver, "sealed", "", now)
	if err != nil {
		t.Fatal(err)
	}
	before := transfer
	if err = transfer.ConfirmRevision("receiver", now, 0); !errors.Is(err, ErrValidation) {
		t.Fatalf("zero revision error = %v, want validation", err)
	}
	if transfer != before {
		t.Fatalf("zero revision mutated transfer: %#v", transfer)
	}
	if err = transfer.ConfirmRevision("receiver", now, 9); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale revision error = %v, want conflict", err)
	}
	if transfer != before {
		t.Fatalf("stale revision mutated transfer: %#v", transfer)
	}
	if err = transfer.ConfirmRevision("receiver", now, before.Revision); err != nil {
		t.Fatal(err)
	}
	if transfer.Status != CustodyConfirmed || transfer.Revision != before.Revision+1 {
		t.Fatalf("valid confirmation = status %s revision %d", transfer.Status, transfer.Revision)
	}
}
