package domain

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSubmitCanceledContextLeavesRequestUnchanged(t *testing.T) {
	now := time.Now().UTC()
	request := NewInspectionRequest("operator", now)
	request.Client.Organization = "lab"
	request.MaterialGrade = "steel"
	request.BatchNumber = "batch"
	request.SampleDescription = "bar"
	request.Requirements = "tensile"
	request.ReceivedAt = now
	before := request
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := request.SubmitContext(ctx, "SMP-20260819-00001", "sample", now)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("submit error = %v", err)
	}
	if !reflect.DeepEqual(request, before) {
		t.Fatalf("request mutated after cancellation: %#v", request)
	}
	if err = request.SubmitContext(context.Background(), "", "sample", now); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty identity error = %v, want validation", err)
	}
	if !reflect.DeepEqual(request, before) {
		t.Fatalf("request mutated after invalid identity: %#v", request)
	}
	zone := time.FixedZone("inspection", 8*60*60)
	submittedAt := time.Date(2026, 8, 19, 12, 30, 0, 0, zone)
	if err = request.SubmitContext(context.Background(), "SMP-20260819-00001", "sample", submittedAt); err != nil {
		t.Fatal(err)
	}
	if request.Reference != "SMP-20260819-00001" || request.SampleID != "sample" || request.Status != RequestSubmitted {
		t.Fatalf("submitted identity/status = %q %q %q", request.Reference, request.SampleID, request.Status)
	}
	if request.UpdatedAt.Location() != time.UTC || !request.UpdatedAt.Equal(submittedAt) {
		t.Fatalf("updatedAt = %v (%v), want UTC instant", request.UpdatedAt, request.UpdatedAt.Location())
	}
}
