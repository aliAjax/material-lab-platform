package attachment

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"material-lab-platform/internal/domain"
)

var errStoreWrite = errors.New("write interrupted")
var errStoreCleanup = errors.New("cleanup unavailable")

type failingStorage struct {
	deleteContextError error
	deleted            bool
}

func (s *failingStorage) Put(context.Context, string, io.Reader) error { return errStoreWrite }
func (s *failingStorage) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("unused")
}
func (s *failingStorage) Delete(ctx context.Context, _ string) error {
	s.deleteContextError = ctx.Err()
	s.deleted = true
	return errStoreCleanup
}

func TestUploadCleansPartialObjectAndPreservesFailureChain(t *testing.T) {
	storage := &failingStorage{}
	service := NewService(storage, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := service.Upload(context.WithoutCancel(ctx), "request", "id", "note.txt", "operator", bytes.NewBufferString("plain text"))
	if !errors.Is(err, errStoreWrite) {
		t.Fatalf("missing write failure: %v", err)
	}
	if !errors.Is(err, errStoreCleanup) {
		t.Fatalf("missing cleanup failure: %v", err)
	}
	if !storage.deleted {
		t.Fatal("partial object was not deleted")
	}
	if storage.deleteContextError != nil {
		t.Fatalf("cleanup used canceled context: %v", storage.deleteContextError)
	}
	if _, err = service.Upload(context.Background(), "", "id", "note.txt", "operator", bytes.NewBufferString("plain text")); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("owner type error = %v", err)
	}
}
