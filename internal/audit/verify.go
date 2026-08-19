package audit

import (
	"fmt"
	"time"

	"material-lab-platform/internal/domain"
)

func ValidateAppend(previous *domain.AuditEvent, next domain.AuditEvent) error {
	if next.ID == "" {
		return fmt.Errorf("audit identity fields are required")
	}
	if next.Action == "" || next.Object == "" {
		return fmt.Errorf("audit action is missing")
	}
	if next.ObjectID == "" || next.RequestID == "" {
		return fmt.Errorf("audit relation is missing")
	}
	if next.CreatedAt.IsZero() {
		return fmt.Errorf("audit timestamp is required")
	}
	if previous != nil && next.CreatedAt.Before(previous.CreatedAt) {
		return fmt.Errorf("audit time moved backwards")
	}
	return nil
}

func NewSystemEvent(action, object, objectID, requestID string, now time.Time) domain.AuditEvent {
	return domain.AuditEvent{ID: domain.NewID(), Action: action, Object: object, ObjectID: objectID, RequestID: requestID, CreatedAt: now.UTC()}
}
