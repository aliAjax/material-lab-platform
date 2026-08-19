package notification

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"material-lab-platform/internal/domain"
)

var ErrAlreadyRead = errors.New("notification already read")

type OwnershipError struct {
	NotificationID string
	Actor          string
}

func (e *OwnershipError) Error() string {
	return fmt.Sprintf("notification %s cannot be read by %s", e.NotificationID, e.Actor)
}
func (e *OwnershipError) Unwrap() error { return domain.ErrForbidden }

func Pending(items []domain.Notification, userID string) []domain.Notification {
	result := make([]domain.Notification, 0)
	for _, item := range items {
		if item.UserID == userID && item.ReadAt.IsZero() {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func MarkRead(item domain.Notification, actor string, now time.Time) (domain.Notification, error) {
	if item.UserID != actor {
		return item, &OwnershipError{NotificationID: item.ID, Actor: actor}
	}
	if !item.ReadAt.IsZero() {
		return item, fmt.Errorf("%w: %s", ErrAlreadyRead, item.ID)
	}
	if now.IsZero() {
		return item, fmt.Errorf("%w: read time is required", domain.ErrValidation)
	}
	item.ReadAt = now.UTC()
	return item, nil
}
