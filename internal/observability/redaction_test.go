package observability

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"material-lab-platform/internal/domain"
	"material-lab-platform/internal/notification"
)

func TestNestedRedactionAndNotificationErrorClassification(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&output, nil)))
	logger.Info("request", slog.Group("request", slog.String("authorization", "Bearer raw"), slog.Group("body", slog.String("refreshToken", "secret-token"))), slog.String("safe", "kept"))
	logged := output.String()
	if strings.Contains(logged, "Bearer raw") || strings.Contains(logged, "secret-token") {
		t.Fatalf("sensitive nested value leaked: %s", logged)
	}
	item := domain.Notification{ID: "notice", UserID: "owner"}
	_, err := notification.MarkRead(item, "other", domain.Notification{}.CreatedAt)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("forbidden chain lost: %v", err)
	}
	var ownership *notification.OwnershipError
	if !errors.As(err, &ownership) {
		t.Fatalf("ownership context lost: %v", err)
	}
	_ = context.Background()
}
