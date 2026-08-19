package audit

import (
	"errors"
	"fmt"
	"time"

	"material-lab-platform/internal/domain"
)

func ValidateAppend(previous *domain.AuditEvent, next domain.AuditEvent) error {
	missing := make([]error, 0)
	for name, value := range map[string]string{
		"id": next.ID, "action": next.Action, "object": next.Object,
		"objectId": next.ObjectID, "requestId": next.RequestID,
	} {
		if value == "" {
			missing = append(missing, fmt.Errorf("%w: audit %s is required", domain.ErrValidation, name))
		}
	}
	if len(missing) != 0 {
		return errors.Join(missing...)
	}
	if next.CreatedAt.IsZero() {
		return fmt.Errorf("%w: audit timestamp is required", domain.ErrValidation)
	}
	if previous != nil && next.CreatedAt.Before(previous.CreatedAt) {
		return fmt.Errorf("%w: audit time moved backwards", domain.ErrConflict)
	}
	return nil
}

func NewSystemEvent(action, object, objectID, requestID string, now time.Time) domain.AuditEvent {
	return domain.AuditEvent{ID: domain.NewID(), Action: action, Object: object, ObjectID: objectID, RequestID: requestID, CreatedAt: now.UTC()}
}
