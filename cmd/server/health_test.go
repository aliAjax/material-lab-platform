package main

import (
	"context"
	"testing"

	"material-lab-platform/internal/health"
)

func TestBuildHealthChecksOmitsUnavailableOptionalDatabase(t *testing.T) {
	checks := buildHealthChecks(nil)
	if _, exists := checks["database"]; exists {
		t.Fatal("nil database check was registered")
	}
	results := health.Run(context.Background(), checks)
	if len(results) != 1 || !results[0].Healthy {
		t.Fatalf("results = %#v", results)
	}
}
