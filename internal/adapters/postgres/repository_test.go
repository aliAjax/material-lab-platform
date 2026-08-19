package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"material-lab-platform/internal/domain"
)

func TestRepositoryRestoresPersistedRequest(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repository, err := NewRepository(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	request := domain.NewInspectionRequest("test", time.Now().UTC())
	request.Client.Organization = "persistence test"
	if err = repository.SaveRequest(request); err != nil {
		t.Fatal(err)
	}
	restored, err := NewRepository(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	value, err := restored.Request(request.ID)
	if err != nil {
		t.Fatal(err)
	}
	if value.Client.Organization != request.Client.Organization {
		t.Fatalf("restored organization = %q", value.Client.Organization)
	}
}
