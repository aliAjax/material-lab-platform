package health

import (
	"context"
	"strings"
	"testing"
)

func TestRunContainsNilAndPanickingChecks(t *testing.T) {
	checks := map[string]Check{
		"nil":   nil,
		"panic": func(context.Context) error { panic("dependency exploded") },
		"ready": func(context.Context) error { return nil },
	}
	results := Run(context.Background(), checks)
	if len(results) != 3 {
		t.Fatalf("results = %d", len(results))
	}
	if results[0].Name != "nil" || results[1].Name != "panic" || results[2].Name != "ready" {
		t.Fatalf("order = %#v", results)
	}
	if results[0].Healthy || !strings.Contains(results[0].Error, "not configured") {
		t.Fatalf("nil result = %#v", results[0])
	}
	if results[1].Healthy || !strings.Contains(results[1].Error, "panicked") {
		t.Fatalf("panic result = %#v", results[1])
	}
	if !results[2].Healthy {
		t.Fatalf("ready result = %#v", results[2])
	}
}
