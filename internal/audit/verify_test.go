package audit

import (
	"errors"
	"testing"
	"time"

	"material-lab-platform/internal/domain"
)

func TestValidateAppendPreservesValidationAndConflictErrors(t *testing.T) {
	err := ValidateAppend(nil, domain.AuditEvent{})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("identity error = %v", err)
	}
	now := time.Now().UTC()
	previous := domain.AuditEvent{CreatedAt: now}
	next := domain.AuditEvent{ID: "id", Action: "save", Object: "sample", ObjectID: "sample-id", RequestID: "request-id", CreatedAt: now.Add(-time.Second)}
	err = ValidateAppend(&previous, next)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ordering error = %v", err)
	}
}
