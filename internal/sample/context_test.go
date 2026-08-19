package sample

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNumberOperationsRejectCanceledContextWithoutPublishingValue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if value, err := FormatNumberContext(ctx, time.Now(), 1); !errors.Is(err, context.Canceled) || value != "" {
		t.Fatalf("format = %q, %v", value, err)
	}
	if day, sequence, err := ParseNumberContext(ctx, "SMP-20260819-00001"); !errors.Is(err, context.Canceled) || !day.IsZero() || sequence != 0 {
		t.Fatalf("parse = %v %d %v", day, sequence, err)
	}
}
