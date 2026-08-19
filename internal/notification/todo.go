package notification

import (
	"sort"
	"time"

	"material-lab-platform/internal/domain"
)

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
		return item, domain.ErrForbidden
	}
	if item.ReadAt.IsZero() {
		item.ReadAt = now.UTC()
	}
	return item, nil
}
